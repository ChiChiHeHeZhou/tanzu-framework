package testutil

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/onsi/ginkgo"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/util/wait"
	yamlutil "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	clusterapiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"
	"sigs.k8s.io/cluster-api/util/secret"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

// CreateResources using unstructured objects from a yaml/json file provided by decoder
func CreateResources(f *os.File, cfg *rest.Config, dynamicClient dynamic.Interface) error {
	var err error
	data, err := os.ReadFile(f.Name())
	if err != nil {
		return err
	}
	decoder := yamlutil.NewYAMLOrJSONDecoder(bytes.NewReader(data), 100)
	mapper, err := apiutil.NewDiscoveryRESTMapper(cfg)
	if err != nil {
		return err
	}
	for {
		resource, unstructuredObj, err := getResource(decoder, mapper, dynamicClient)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return err
			}
		}
		_, err = resource.Create(context.Background(), unstructuredObj, metav1.CreateOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}

// DeleteResources using unstructured objects from a yaml/json file provided by decoder
func DeleteResources(f *os.File, cfg *rest.Config, dynamicClient dynamic.Interface, waitForDeletion bool) error {
	var err error
	data, err := os.ReadFile(f.Name())
	if err != nil {
		return err
	}
	decoder := yamlutil.NewYAMLOrJSONDecoder(bytes.NewReader(data), 100)
	mapper, err := apiutil.NewDiscoveryRESTMapper(cfg)
	if err != nil {
		return err
	}
	for {
		resource, unstructuredObj, err := getResource(decoder, mapper, dynamicClient)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return err
			}
		}
		deletionPropagation := metav1.DeletePropagationForeground
		gracePeriodSeconds := int64(0)
		if err := resource.Delete(context.Background(), unstructuredObj.GetName(),
			metav1.DeleteOptions{GracePeriodSeconds: &gracePeriodSeconds,
				PropagationPolicy: &deletionPropagation}); err != nil {
			return err
		}

	}
	if waitForDeletion {
		// verify deleted
		data, err = os.ReadFile(f.Name())
		if err != nil {
			return err
		}
		decoder := yamlutil.NewYAMLOrJSONDecoder(bytes.NewReader(data), 100)
		mapper, err = apiutil.NewDiscoveryRESTMapper(cfg)
		if err != nil {
			return err
		}
		for {
			resource, unstructuredObj, err := getResource(decoder, mapper, dynamicClient)
			if err != nil {
				if err == io.EOF {
					break
				} else {
					return err
				}
			}
			fmt.Fprintln(ginkgo.GinkgoWriter, "wait for deletion", unstructuredObj.GetName())
			if err := wait.Poll(time.Second*5, time.Second*10, func() (done bool, err error) {
				obj, err := resource.Get(context.Background(), unstructuredObj.GetName(), metav1.GetOptions{})

				if err == nil {
					fmt.Fprintln(ginkgo.GinkgoWriter, "remove finalizers", obj.GetFinalizers(), unstructuredObj.GetName())
					obj.SetFinalizers(nil)
					_, err = resource.Update(context.Background(), obj, metav1.UpdateOptions{})
					if err != nil {
						return false, err
					}
					return false, nil
				}
				if apierrors.IsNotFound(err) {
					return true, nil
				}
				return false, err

			}); err != nil {
				return err
			}
		}
	}

	return nil
}

//CreateKubeconfigSecret create a secret with kubeconfig token for the cluster provided by client
func CreateKubeconfigSecret(cfg *rest.Config, clusterName string, namespace string, client client.Client) error {
	clusters := make(map[string]*clientcmdapi.Cluster)
	clusters[clusterName] = &clientcmdapi.Cluster{
		Server:                   cfg.Host,
		CertificateAuthorityData: cfg.CAData,
	}
	contexts := make(map[string]*clientcmdapi.Context)
	contexts["default-context"] = &clientcmdapi.Context{
		Cluster:   clusterName,
		Namespace: namespace,
		AuthInfo:  "default",
	}
	authinfos := make(map[string]*clientcmdapi.AuthInfo)
	authinfos["default"] = &clientcmdapi.AuthInfo{
		Token: cfg.BearerToken,
	}
	clientConfig := clientcmdapi.Config{
		Kind:           "Config",
		APIVersion:     "v1",
		Clusters:       clusters,
		Contexts:       contexts,
		CurrentContext: "default-context",
		AuthInfos:      authinfos,
	}
	kubeconfig, err := clientcmd.Write(clientConfig)
	if err != nil {
		return err
	}
	kc := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secret.Name(clusterName, secret.Kubeconfig),
			Namespace: namespace,
			Labels: map[string]string{
				clusterapiv1alpha3.ClusterLabelName: clusterName,
			},
		},
		Data: map[string][]byte{
			secret.KubeconfigDataName: kubeconfig,
		},
		Type: clusterapiv1alpha3.ClusterSecretType,
	}

	return client.Create(context.Background(), kc)
}

func getResource(decoder *yamlutil.YAMLOrJSONDecoder, mapper meta.RESTMapper, dynamicClient dynamic.Interface) (
	dynamic.ResourceInterface, *unstructured.Unstructured, error) {
	var rawObj runtime.RawExtension
	if err := decoder.Decode(&rawObj); err != nil {
		return nil, nil, err
	}

	obj, gvk, err := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme).Decode(rawObj.Raw, nil, nil)
	if err != nil {
		return nil, nil, err
	}

	unstructuredMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return nil, nil, err
	}

	unstructuredObj := &unstructured.Unstructured{Object: unstructuredMap}

	mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return nil, nil, err
	}

	var resource dynamic.ResourceInterface
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		if unstructuredObj.GetNamespace() == "" {
			unstructuredObj.SetNamespace("default")
		}
		resource = dynamicClient.Resource(mapping.Resource).Namespace(unstructuredObj.GetNamespace())
	} else {
		resource = dynamicClient.Resource(mapping.Resource)
	}
	return resource, unstructuredObj, nil
}
