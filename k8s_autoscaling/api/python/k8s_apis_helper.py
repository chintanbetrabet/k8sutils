import json
import urlparse
import requests


def make_kubernetes_api_get_request(uri):
    base_url = "https://kubernetes.default.svc.cluster.local"
    url = urlparse.urljoin(base_url, uri)

    ca_cert_file = "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"

    auth_token_file = "/var/run/secrets/kubernetes.io/serviceaccount/token"
    auth_token = None
    with open(auth_token_file) as fptr:
        auth_token = fptr.read()

    headers = {'Authorization': "Bearer %s" % auth_token}
    response = requests.get(url, verify=ca_cert_file, headers=headers)
    return response


def k8s_patch_req(uri, payload):
    base_url = "https://kubernetes.default.svc.cluster.local"
    url = urlparse.urljoin(base_url, uri)

    ca_cert_file = "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"

    auth_token_file = "/var/run/secrets/kubernetes.io/serviceaccount/token"
    auth_token = None
    with open(auth_token_file) as fptr:
        auth_token = fptr.read()

    headers = {'Authorization': "Bearer %s" % auth_token,
               'Content-Type': 'application/json-patch+json', 'Accept': 'application/json'}
    response = requests.patch(url, verify=ca_cert_file,
                              headers=headers, data=payload)
    return response


def pods(namespace):
    return json.loads(make_kubernetes_api_get_request("/api/v1/namespaces/%s/pods" % namespace).content)["items"]


def deployment_data(namespace, deployment):
    return json.loads(make_kubernetes_api_get_request("/apis/apps/v1/namespaces/%s/deployments/%s" % (namespace, deployment)).content)


def pod_data(namespace, pod_name):
    return json.loads(make_kubernetes_api_get_request("/api/v1/namespaces/%s/pods/%s" % (namespace, pod_name)).content)


def pod_labels(namespace, pod_name):
    return pod_data(namespace, pod_name)["metadata"].get("labels", {})


def deployment_pod_labels(namespace, deployment):
    return deployment_data(namespace, deployment).get("spec", {}).get("template", {}).get("metadata", {}).get("labels", {})


def add_label(key, value, pod_name, namespace):
    payload = '[ { "op": "add", "path": "/metadata/labels/%s", "value": "%s" } ]' % (
        key, value)
    return k8s_patch_req("/api/v1/namespaces/%s/pods/%s" % (namespace, pod_name), payload)


def scale(deployment, replicas, namespace):
    payload = '[{ "op": "replace", "path": "/spec/replicas", "value": %s }]' % replicas
    return k8s_patch_req("/apis/apps/v1/namespaces/%s/deployments/%s" % (namespace, deployment), payload)


def pods_with_labels(namespace, labels_map):
    uri = "/api/v1/namespaces/%s/pods" % namespace
    labels_query_strings = []
    for key in labels_map:
        labels_query_strings.append(key + "%3D" + labels_map[key])
    query = "?labelSelector=" + ",".join(labels_query_strings)
    print json.loads(make_kubernetes_api_get_request(uri + query).content)
    return json.loads(make_kubernetes_api_get_request(uri + query).content)["items"]


def pod_names_with_labels(namespace, labels_map):
    pods_data = pods_with_labels(namespace, labels_map)
    names = []
    for data in pods_data:
        names.append(data["metadata"]["name"])
    return names


def check_pods_exist(expected_pods, present_pods):
    passed = all(pod in present_pods for pod in expected_pods)
    if passed is True:
        print "PASS"
    else:
        for pod in expected_pods:
            if pod not in present_pods:
                print "%s missing" % pod
        print "FAIL"
map = {}
map[' app'] = 'nginx'
#print pod_names_with_labels('test-rbac', map)
print scale('nginx', 4, 'async')