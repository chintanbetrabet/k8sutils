'''
1. Identify target pods to preserve
2. Change their label to be outside of deployment -> new pods will spawn in their place [ possibly all existing pods also ]
3. Downscale the deployment to max(0, desired_scale - exempted_pods), then upscale to desired_scale
4. Add back these pods to deployment

 Case 1: Number of exempted pods <= New scale count -> Tested
 -> Some 'allowed to die' pods will die and 'exempted' pods take their place
 -> Number of pods in system ==  New scale count

 Case 2: Number of exempted pods > New scale count -> Tested
 -> Other pods will die off and some of these pods will in GS, rest will serve traffic
 -> All pods in the system are 'exempted'
 -> Number of pods in system ==  Number of exempted pods
 -> Number of exempted pods - New scale count are extra in the system
'''

import time
import copy
import multiprocessing
from joblib import Parallel, delayed
from k8s_apis_helper import add_label, scale, pod_names_with_labels


def detach_pod_with_labels(pod, deployment_labels, namespace):
    for key in deployment_labels:
        add_label(key=key, value=str(deployment_labels[key]) +
                  "-detached", pod_name=pod, namespace=namespace)


def attach_pod_with_labels(pod, deployment_labels, namespace):
    for key in deployment_labels:
        add_label(key=key, value=str(deployment_labels[key]),
                  pod_name=pod, namespace=namespace)


def detach_pods(pods, deployment_labels, namespace):
    num_cores = multiprocessing.cpu_count()
    Parallel(n_jobs=num_cores)(delayed(detach_pod_with_labels)(
        pod=pod, deployment_labels=deployment_labels, namespace=namespace) for pod in pods)

    while any(pod in pod_names_with_labels(namespace=namespace, labels_map=deployment_labels) for pod in pods):
        time.sleep(1)


def attach_pods(pods, deployment_labels, namespace):
    num_cores = multiprocessing.cpu_count()
    Parallel(n_jobs=num_cores)(delayed(attach_pod_with_labels)(
        pod=pod, deployment_labels=deployment_labels, namespace=namespace) for pod in pods)
    while any(pod not in pod_names_with_labels(namespace=namespace, labels_map=deployment_labels) for pod in pods):
        time.sleep(1)


def compute_exempted_pods(deployment_labels):
    custom_labels = copy.deepcopy(deployment_labels)
    custom_labels["termination-candidate"] = "false"
    return pod_names_with_labels("default", custom_labels)


def exempt_pods_from_downscale(exempted_pods, namespace):
    for pod in exempted_pods:
        add_label("termination-candidate", "false", pod, namespace)


def downscale(target_scale, deployment_name, deployment_labels, namespace):
    exempted_pods = compute_exempted_pods(deployment_labels)
    number_of_exempted_pods = len(exempted_pods)
    detach_pods(pods=exempted_pods,
                deployment_labels=deployment_labels, namespace=namespace)
    scale(namespace=namespace, deployment=deployment_name,
          replicas=max(target_scale - number_of_exempted_pods, 0))
    scale(deployment=deployment_name, replicas=target_scale, namespace=namespace)
    attach_pods(pods=exempted_pods,
                deployment_labels=deployment_labels, namespace=namespace)
