/*
Copyright 2016 The Kubernetes Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Note: the example only works with the code within the same release/branch.
package main

import (
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
//	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
  "k8s.io/client-go/rest"
)

func main() {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
  client, err := dynamic.NewForConfig(config)
  	if err != nil {
  		panic(err)
  	}
    namespace := "default"
    deploymentRes := schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}

  	// deployment := &unstructured.Unstructured{
  	// 	Object: map[string]interface{}{
  	// 		"apiVersion": "apps/v1",
  	// 		"kind":       "Deployment",
  	// 		"metadata": map[string]interface{}{
  	// 			"name": "demo-deployment",
  	// 		},
  	// 		"spec": map[string]interface{}{
  	// 			"replicas": 2,
  	// 			"selector": map[string]interface{}{
  	// 				"matchLabels": map[string]interface{}{
  	// 					"app": "demo",
  	// 				},
  	// 			},
  	// 			"template": map[string]interface{}{
  	// 				"metadata": map[string]interface{}{
  	// 					"labels": map[string]interface{}{
  	// 						"app": "demo",
  	// 					},
  	// 				},
		// 
  	// 				"spec": map[string]interface{}{
  	// 					"containers": []map[string]interface{}{
  	// 						{
  	// 							"name":  "web",
  	// 							"image": "nginx:1.12",
  	// 							"ports": []map[string]interface{}{
  	// 								{
  	// 									"name":          "http",
  	// 									"protocol":      "TCP",
  	// 									"containerPort": 80,
  	// 								},
  	// 							},
  	// 						},
  	// 					},
  	// 				},
  	// 			},
  	// 		},
  	// 	},
  	// }

  	// Create Deployment

		
		deployment, err := client.Resource(deploymentRes).Namespace(namespace).Get("demo-deployment", metav1.GetOptions{})
		fmt.Printf("%q.\n", deployment)
		
		// fmt.Println("Creating deployment...")
		// result, err := client.Resource(deploymentRes).Namespace(namespace).Create(deployment, metav1.CreateOptions{})
		// if err != nil {
		// 	panic(err)
		// }
		// fmt.Printf("Created deployment %q.\n", result.GetName())
		
	// for {
	// 	// get pods in all the namespaces by omitting namespace
	// 	// Or specify namespace to get pods in particular namespace
	// 	pods, err := clientset.AppsV1Client().Deployments("test-rbac").List(metav1.ListOptions{})
	// 	if err != nil {
	// 		panic(err.Error())
	// 	}
	// 	fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))
  // 
	// 	// Examples for error handling:
	// 	// - Use helper functions e.g. errors.IsNotFound()
	// 	// - And/or cast to StatusError and use its properties like e.g. ErrStatus.Message
	// 	//_, err = clientset.CoreV1().Deployments("test-rbac")
	// // 	if errors.IsNotFound(err) {
	// // 		fmt.Printf("Pod example-xxxxx not found in default namespace\n")
	// // 	} else if statusError, isStatus := err.(*errors.StatusError); isStatus {
	// // 		fmt.Printf("Error getting pod %v\n", statusError.ErrStatus.Message)
	// // 	} else if err != nil {
	// // 		panic(err.Error())
	// // 	} else {
	// // 		fmt.Printf("Found example-xxxxx pod in default namespace\n")
	// // 	}
  // // 
	// // 	time.Sleep(10 * time.Second)
	// }
}
