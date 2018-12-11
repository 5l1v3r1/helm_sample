package main

import (
	"fmt"
	"log"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/helm/portforwarder"
	"k8s.io/helm/pkg/kube"
)

var (
	tillerTunnel *kube.Tunnel
	tillerHost   string
)

const (
	ReleaseName = "hoge-release"
)

func main() {
	err := setupConnection()
	if err != nil {
		log.Println(err)
	}
	client := newClient()
	opts := []helm.DeleteOption{
		helm.DeleteDisableHooks(false),
		helm.DeletePurge(true),
		helm.DeleteTimeout(10),
	}
	res, err := client.DeleteRelease(ReleaseName, opts...)
	if err != nil {
		log.Print(err)
	}
	fmt.Printf("%+v", res)
}
func newClient() helm.Interface {
	options := []helm.Option{
		helm.Host(tillerHost),
		helm.ConnectTimeout(10),
	}
	return helm.NewClient(options...)
}
func getKubeClient(context string, kubeconfig string) (*rest.Config, kubernetes.Interface, error) {
	config, err := configForContext(context, kubeconfig)
	if err != nil {
		return nil, nil, err
	}
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, fmt.Errorf("could not get Kubernetes client: %s", err)
	}
	return config, client, nil
}
func configForContext(context string, kubeconfig string) (*rest.Config, error) {
	config, err := GetConfig(context, kubeconfig).ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("could not get Kubernetes config for context %q: %s", context, err)
	}
	return config, nil
}

// Port Forward
func setupConnection() error {
	config, client, err := getKubeClient("k8s.cluster.domain", "~/.kube/config") // change your home dir
	if err != nil {
		return err
	}
	tillerTunnel, err := portforwarder.New("kube-system", client, config)
	if err != nil {
		return err
	}
	tillerHost = fmt.Sprintf("127.0.0.1:%d", tillerTunnel.Local)
	log.Printf("Created tunnel using local port: '%d'\n", tillerTunnel.Local)
	return nil
}
func GetConfig(context string, kubeconfig string) clientcmd.ClientConfig {
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	rules.DefaultClientConfig = &clientcmd.DefaultClientConfig
	overrides := &clientcmd.ConfigOverrides{ClusterDefaults: clientcmd.ClusterDefaults}
	if context != "" {
		overrides.CurrentContext = context
	}
	if kubeconfig != "" {
		rules.ExplicitPath = kubeconfig
	}
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, overrides)
}
