package main

import (
	"fmt"
	"log"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

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
	kubeConfig, err := rest.InClusterConfig()
	if err != nil {
		log.Print("Failed to get kubernetes in cluster config %+v", err)
		return
	}
	kubeClientSet, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		log.Printf("failed to construct kubernetes client set from config: %+v", err)
		return
	}

	tillerTunnel, err := portforwarder.New("kube-system", kubeClientSet, kubeConfig)
	if err != nil {
		log.Print(err, "failed to portforward")
		return
	}
	tillerHost = fmt.Sprintf("127.0.0.1:%d", tillerTunnel.Local)

	helmOpts := []helm.Option{
		helm.Host(tillerHost),
		helm.ConnectTimeout(10), // デフォルトのタイムアウト値が0秒なため必須
	}
	helmClient := helm.NewClient(helmOpts...)
	listOpts := []helm.ReleaseListOption{
		helm.ReleaseListFilter(ReleaseName),
	}

	listRelease, err := helmClient.ListReleases(listOpts...)
	if err != nil {
		log.Print(err, "failed to get ReleaseName")
		return
	}

	if listRelease == nil {
		log.Printf("%s not found", ReleaseName)
		return
	}

	deleteOpts := []helm.DeleteOption{
		helm.DeleteDisableHooks(false),
		helm.DeletePurge(true),
		helm.DeleteTimeout(10), // デフォルトのタイムアウト値が0秒なため必須
	}

	_, err = helmClient.DeleteRelease(ReleaseName, deleteOpts...)
	if err != nil {
		log.Print("failed to delete")
		return
	}

}
