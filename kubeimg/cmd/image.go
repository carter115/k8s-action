package cmd

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubeimg/pkg/kube"
	"kubeimg/pkg/mtable"
)

var imageCmd = &cobra.Command{
	Use:   "image",
	Short: "show image",
	Run:   image,
}
var title = []string{"NAMESPACE", "TYPE", "NAME", "CNAME", "IMAGE"}

func image(cmd *cobra.Command, args []string) {
	client, err := kube.ClientSet(kubeconfig)
	if err != nil {
		fmt.Println(err)
		return
	}
	//fmt.Printf("cmd: %+v\n",cmd)
	//fmt.Printf("args: %+v\n", args)
	ns, _ := cmd.Flags().GetString("namespace")
	//fmt.Println("namespace:", ns)

	// 全局资源列表
	var resourceList []interface{}
	if flag, _ := cmd.Flags().GetBool("deployments"); flag {
		resList, err := client.AppsV1().Deployments(ns).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			fmt.Printf("list deployments error: %v\n", err)
		}
		resourceList = append(resourceList, resList)
	}

	if flag, _ := cmd.Flags().GetBool("daemonsets"); flag {
		resList, err := client.AppsV1().DaemonSets(ns).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			fmt.Printf("list daemonsets error: %v\n", err)
		}
		resourceList = append(resourceList, resList)
	}

	if flag, _ := cmd.Flags().GetBool("statefulsets"); flag {
		resList, err := client.AppsV1().StatefulSets(ns).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			fmt.Printf("list statefulsets error: %v\n", err)
		}
		resourceList = append(resourceList, resList)
	}

	if flag, _ := cmd.Flags().GetBool("jobs"); flag {
		resList, err := client.BatchV1().Jobs(ns).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			fmt.Printf("list jobs error: %v\n", err)
		}
		resourceList = append(resourceList, resList)
	}

	if flag, _ := cmd.Flags().GetBool("cronjobs"); flag {
		resList, err := client.BatchV1().CronJobs(ns).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			fmt.Printf("list cronjobs error: %v\n", err)
		}
		resourceList = append(resourceList, resList)
	}
	//fmt.Println("resourceList", len(resourceList), resourceList)

	// 资源的镜像信息
	var resourceMapList = []map[string]string{}
	for _, v := range resourceList {
		switch t := v.(type) {
		case *v1.DeploymentList:
			var resList = make(map[string]string, 0)
			for _, res := range t.Items {
				containers := res.Spec.Template.Spec.Containers
				for j := 0; j < len(containers); j++ {
					resList[title[0]] = ns
					resList[title[1]] = "deployment"
					resList[title[2]] = res.Name
					resList[title[3]] = containers[j].Name
					resList[title[4]] = containers[j].Image
					resourceMapList = append(resourceMapList, resList)
				}
			}
		case *v1.DaemonSetList:
			var resList = make(map[string]string, 0)
			for _, res := range t.Items {
				containers := res.Spec.Template.Spec.Containers
				for j := 0; j < len(containers); j++ {
					resList[title[0]] = ns
					resList[title[1]] = "daemonset"
					resList[title[2]] = res.Name
					resList[title[3]] = containers[j].Name
					resList[title[4]] = containers[j].Image
					resourceMapList = append(resourceMapList, resList)
				}
			}
		case *v1.StatefulSetList:
			var resList = make(map[string]string, 0)
			for _, res := range t.Items {
				containers := res.Spec.Template.Spec.Containers
				for j := 0; j < len(containers); j++ {
					resList[title[0]] = ns
					resList[title[1]] = "statefulset"
					resList[title[2]] = res.Name
					resList[title[3]] = containers[j].Name
					resList[title[4]] = containers[j].Image
					resourceMapList = append(resourceMapList, resList)
				}
			}
		case *batchv1.JobList:
			var resList = make(map[string]string, 0)
			for _, res := range t.Items {
				containers := res.Spec.Template.Spec.Containers
				for j := 0; j < len(containers); j++ {
					resList[title[0]] = ns
					resList[title[1]] = "job"
					resList[title[2]] = res.Name
					resList[title[3]] = containers[j].Name
					resList[title[4]] = containers[j].Image
					resourceMapList = append(resourceMapList, resList)
				}
			}
		case *batchv1.CronJobList:
			var resList = make(map[string]string, 0)
			for _, res := range t.Items {
				containers := res.Spec.JobTemplate.Spec.Template.Spec.Containers
				for j := 0; j < len(containers); j++ {
					resList[title[0]] = ns
					resList[title[1]] = "cronjob"
					resList[title[2]] = res.Name
					resList[title[3]] = containers[j].Name
					resList[title[4]] = containers[j].Image
				}
			}
		}
	}

	mtable.GenTable(resourceMapList,title).PrintTable()
}



func init() {
	rootCmd.AddCommand(imageCmd)

	imageCmd.Flags().BoolP("deployments", "d", false, "show deployments image")
	imageCmd.Flags().BoolP("daemonsets", "e", false, "show daemonsets image")
	imageCmd.Flags().BoolP("statefulsets", "f", false, "show statefulsets image")
	imageCmd.Flags().BoolP("jobs", "o", false, "show jobs image")
	imageCmd.Flags().BoolP("cronjobs", "b", false, "show cronjobs image")
	imageCmd.Flags().BoolP("json", "j", false, "show json format")
}
