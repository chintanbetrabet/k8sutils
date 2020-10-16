import requests
import json
import urlparse

def deployments_to_monitor():
    pass

def deployment_data():
  # Min/max
  # Metrics description
  # autoscaling logic
  pass

def collect_cpu_metrics():
    pass

def collect_memory_metrics():
    pass

def collect_statsd_metrics():
    pass

def collect_metrics(deployment):
    pass

def compute_desired_replicas(deployment):
    autoscaling_data = deployment_data()
    pass

def current_replicas(deployment):
    pass

def scale(deployment, replicas):
  if replicas >= current_replicas(deployment):
    upscale(deployment, replicas)
  else:
    downscale(deployment, replicas)

def upscale(deployment, replicas):
    pass

def downscale(deployment, replicas):
    pass

def candidates_for_downscale(deployment, replicas):
    pass

def detach_pods_from_deployment(deployment:, identifying_label:, label_value:, pod_list:):
    pass

def patch_api_call():
    pass

def k8s_patch_req(uri, payload):
    base_url = "https://kubernetes.default.svc.cluster.local"
    url = urlparse.urljoin(base_url, uri)
    
    ca_cert_file = "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"

    auth_token_file = "/var/run/secrets/kubernetes.io/serviceaccount/token"
    auth_token = None
    with open(auth_token_file) as fptr:
        auth_token = fptr.read()

    headers={ 'Authorization': "Bearer %s" % auth_token, 'Content-Type': 'application/json-patch+json', 'Accept' : 'application/json' }
    response = requests.patch(url, verify=ca_cert_file, headers=headers, data=payload)
    
    return response

def scale_api_call(deployment, namespace, scale):
    payload = '[{ "op": "replace", "path": "/spec/replicas", "value": %s }]' % scale
    return k8s_patch_req("/apis/apps/v1/namespaces/%s/deployments/%s" % (deployment, namespace) payload)
    #return make_kubernetes_api_request("/api/v1/namespaces/test-rbac/pods/nginx-deployment-66f7f56f56-4xklv/log")