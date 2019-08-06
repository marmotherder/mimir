package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/marmotherder/mimir/clients"

	"k8s.io/api/admission/v1beta1"
	core_v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

//patchReq is a struct for a patch request that doesn't exist in types from admission
type patchReq struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

// hook is the main function for inbound hook requests
func hook(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "invalid Content-Type, expect `application/json`", http.StatusUnsupportedMediaType)
		return
	}

	as := v1beta1.AdmissionResponse{}
	ar, pod, err := readRequest(r.Body)

	if err != nil {
		setResultMessage(&as, err.Error())
	}

	if ar != nil && ar.Request.Operation == v1beta1.Create && err == nil {
		if err != nil {
			setResultMessage(&as, err.Error())
		} else {
			err := runCreateHook(w, r, ar, pod, &as)
			if err != nil {
				setResultMessage(&as, err.Error())
			} else {
				as.Allowed = true
			}
		}
	} else if ar != nil && ar.Request.Operation == v1beta1.Delete && err == nil {
		as.Allowed = true
		kc, err := clients.NewK8SClient(opts.IsPod, opts.KubeconfigPath)
		if err != nil {
			setResultMessage(&as, err.Error())
		} else {
			genName := fmt.Sprintf("%s-%s", release, ar.Request.Name)
			err := kc.CoreV1().Secrets(ar.Request.Namespace).Delete(genName, &meta_v1.DeleteOptions{})
			if err != nil {
				setResultMessage(&as, err.Error())
			} else {
				setResultMessage(&as, fmt.Sprintf("Deleted secret %s in namespace %s", genName, ar.Request.Namespace))
			}
		}
	}

	dispatchResponse(ar, as, w)
}

// loadSecret will load a secret from the remote secrets manager based on the annotations on the pod
func loadSecret(name, namespace, remote string) (*core_v1.Secret, *kubernetes.Clientset, map[string]string, error) {
	smc, _, err := loadClient()
	if err != nil {
		return nil, nil, nil, err
	}

	secret, err := smc.GetSecret(remote)
	if err != nil {
		return nil, nil, nil, err
	}

	kc, err := clients.NewK8SClient(opts.IsPod, opts.KubeconfigPath)

	genName := fmt.Sprintf("%s-%s", release, name)

	kc.CoreV1().Secrets(namespace).Delete(genName, &meta_v1.DeleteOptions{})

	data := make(map[string][]byte)
	for k, v := range secret.Data {
		data[k] = []byte(v)
	}

	k8sSecret := &core_v1.Secret{
		Type: core_v1.SecretTypeOpaque,
		Data: data,
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      genName,
			Namespace: namespace,
		},
	}

	k8sSecret.Annotations = map[string]string{clients.Hook: sOpts.Hook}

	return k8sSecret, kc, secret.Data, nil
}

// addPatchReq adds a patchReq struct to a slice in a repeatable generic way
func addPatchReq(isNew func() bool, op string, path string, patchReqs *[]patchReq, value interface{}) {
	req := patchReq{
		Op:    op,
		Value: value,
	}

	if isNew() {
		req.Path = path
	} else {
		req.Path = fmt.Sprintf("%s/-", path)
	}
	*patchReqs = append(*patchReqs, req)
}

// runCreateHook is a split out function for running the steps when creating a pod
func runCreateHook(w http.ResponseWriter, r *http.Request, ar *v1beta1.AdmissionReview, pod *core_v1.Pod, as *v1beta1.AdmissionResponse) error {
	remote, path, local, runEnv, err := shouldMutate(pod)
	if err != nil {
		return err
	}
	if remote == nil {
		return nil
	}

	podName := pod.Name
	if local != nil && *local != "" {
		podName = *local
	}
	if podName == "" {
		podName = string(ar.Request.UID)
	}

	namespace := pod.Namespace
	if namespace == "" {
		namespace = ar.Request.Namespace
	}
	if namespace == "" {
		namespace = "default"
	}

	k8sSecret, kc, data, err := loadSecret(podName, namespace, *remote)
	if err != nil {
		return err
	}

	patchReqs := make([]patchReq, 0)

	if path != nil {
		vol := core_v1.Volume{Name: k8sSecret.Name}
		vol.Secret = &core_v1.SecretVolumeSource{SecretName: k8sSecret.Name}
		addPatchReq(func() bool { return len(pod.Spec.Volumes) == 0 }, "add", "/spec/volumes", &patchReqs, vol)
	}

	envs := make([]core_v1.EnvVar, 0)
	if runEnv {
		for k := range data {
			selector := &core_v1.SecretKeySelector{Key: k}
			selector.Name = k8sSecret.Name
			envs = append(envs, core_v1.EnvVar{
				Name: k,
				ValueFrom: &core_v1.EnvVarSource{
					SecretKeyRef: selector,
				},
			})
		}
	}

	containerPatches := func(specPath string, containers []core_v1.Container) {
		for idx, container := range containers {
			if path != nil {
				add := core_v1.VolumeMount{
					Name:      k8sSecret.Name,
					ReadOnly:  true,
					MountPath: *path,
				}
				addPatchReq(func() bool { return len(container.VolumeMounts) == 0 }, "add", fmt.Sprintf("/spec/%s/%d/volumeMounts", specPath, idx), &patchReqs, add)
			}
			if len(envs) > 0 {
				addPatchReq(func() bool { return len(container.Env) == 0 }, "add", fmt.Sprintf("/spec/%s/%d/env", specPath, idx), &patchReqs, envs)
			}
		}
	}

	containerPatches("containers", pod.Spec.Containers)
	containerPatches("initContainers", pod.Spec.InitContainers)

	annotations := make(map[string]string)
	if len(pod.Annotations) > 0 {
		annotations = pod.Annotations
	}
	annotations[clients.Managed] = "true"
	addPatchReq(func() bool { return true }, "add", "/metadata/annotations", &patchReqs, annotations)

	patchData, err := json.Marshal(patchReqs)
	if err != nil {
		return err
	}

	as.Patch = patchData
	as.PatchType = func() *v1beta1.PatchType {
		p := v1beta1.PatchTypeJSONPatch
		return &p
	}()

	if _, err := kc.CoreV1().Secrets(namespace).Create(k8sSecret); err != nil {
		return err
	}
	setResultMessage(as, fmt.Sprintf("Created secret %s in namespace %s", k8sSecret.Name, namespace))
	return nil
}

// readRequest reads the inbound hook request
func readRequest(body io.ReadCloser) (*v1beta1.AdmissionReview, *core_v1.Pod, error) {
	var ar v1beta1.AdmissionReview
	if err := json.NewDecoder(body).Decode(&ar); err != nil {
		return nil, nil, err
	}
	if len(ar.Request.Object.Raw) > 0 {
		var pod core_v1.Pod
		if err := json.Unmarshal(ar.Request.Object.Raw, &pod); err != nil {
			return nil, nil, err
		}
		return &ar, &pod, nil
	}
	return &ar, nil, nil
}

// shouldMutate will scan the pod for the desired annotations
// If the annotations exist, return the paths we need to progress with the secrets creation
func shouldMutate(pod *core_v1.Pod) (*string, *string, *string, bool, error) {

	isHook := false
	remote := ""
	path := ""
	env := "false"
	local := ""

	for k, v := range pod.Annotations {
		switch k {
		case clients.Hook:
			if v == sOpts.Hook {
				isHook = true
			}
		case clients.Remote:
			remote = v
		case clients.Path:
			path = v
		case clients.Env:
			env = v
		case clients.Local:
			local = v
		}
	}

	runEnv, _ := strconv.ParseBool(env)

	if !isHook {
		return nil, nil, nil, false, nil
	}

	if remote == "" {
		return nil, nil, nil, false, errors.New("Missing properties for remote secret name")
	}

	if path != "" {
		return &remote, &path, &local, runEnv, nil
	}
	return &remote, nil, &local, runEnv, nil
}

// setResultMessage adds a message to the response struct, can be errors or otherwise
func setResultMessage(as *v1beta1.AdmissionResponse, message string) {
	status := meta_v1.Status{Message: message}
	as.Result = &status
}

// dispatchResponse writes out the json response payload from the hook
func dispatchResponse(ar *v1beta1.AdmissionReview, as v1beta1.AdmissionResponse, w http.ResponseWriter) {
	resp := v1beta1.AdmissionReview{}
	resp.Response = &as
	if ar.Request != nil {
		resp.Response.UID = ar.Request.UID
	}
	payloadjson, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "could not decode request message", http.StatusInternalServerError)
	}
	log.Println(string(payloadjson))
	w.Write(payloadjson)
}

// shutdownServer runs when a term or kill command reaches the application. It is designed to deconstruct the k8s objects created by the init container
func shutdownServer(srv *http.Server) {
	log.Println("Server shutdown started. Will try to cleanup dynamic resources")

	kc, err := clients.NewK8SClient(opts.IsPod, opts.KubeconfigPath)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Using release %s\n", release)

	whs, err := kc.AdmissionregistrationV1beta1().MutatingWebhookConfigurations().List(meta_v1.ListOptions{LabelSelector: fmt.Sprintf("app=%s", release)})
	if err != nil {
		log.Println(err.Error())
	} else {
		for _, whs := range whs.Items {
			log.Printf("Removing mutating webhook configuration %s\n", whs.Name)
			kc.AdmissionregistrationV1beta1().MutatingWebhookConfigurations().Delete(whs.Name, &meta_v1.DeleteOptions{})
		}
	}

	csr, err := kc.CertificatesV1beta1().CertificateSigningRequests().List(meta_v1.ListOptions{LabelSelector: fmt.Sprintf("app=%s", release)})
	if err != nil {
		log.Println(err.Error())
	} else {
		for _, csr := range csr.Items {
			log.Printf("Removing csr %s\n", csr.Name)
			kc.CertificatesV1beta1().CertificateSigningRequests().Delete(csr.Name, &meta_v1.DeleteOptions{})
		}
	}

	srv.Shutdown(context.Background())
}
