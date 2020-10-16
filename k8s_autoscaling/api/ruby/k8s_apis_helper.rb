require 'json'
require 'rest-client'
require 'uri'


def make_kubernetes_api_get_request(uri)
    base_url = "https://kubernetes.default.svc.cluster.local"
    url = (base_url +  uri)

    ca_cert_file = "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"

    auth_token_file = "/var/run/secrets/kubernetes.io/serviceaccount/token"
    auth_token = File.read(auth_token_file)
    headers = {'Authorization'=> "Bearer #{auth_token}"}
    RestClient::Request.execute(method: :get, url: url,
                            ssl_ca_file: ca_cert_file,
                            headers: headers
                          )
end

def pods(namespace)
    JSON.parse(make_kubernetes_api_get_request("/api/v1/namespaces/#{namespace}/pods").body)["items"]
end

def deployment_data(namespace, deployment)
    JSON.parse(make_kubernetes_api_get_request("/apis/apps/v1/namespaces/#{namespace}/deployments/#{deployment}").body)
end

def pod_data(namespace, pod_name)
    JSON.parse(make_kubernetes_api_get_request("/api/v1/namespaces/#{namespace}/pods/#{pod_name}").body)
end

def pod_labels(namespace, pod_name)
    pod_data(namespace, pod_name)["metadata"]["labels"] ||  {}
end

def deployment_pod_labels(namespace, deployment)
    return deployment_data(namespace, deployment).fetch("spec", {}).fetch("template", {}).fetch("metadata", {}).fetch("labels", {})
end

def pods_with_labels(namespace, labels_map)
    uri = "/api/v1/namespaces/#{namespace}/pods"
    labels_query_strings = []
    labels_map.keys.each do |key|
        labels_query_strings + (key + "%3D" + labels_map[key])
    end
    query = "?labelSelector=" + ",".join(labels_query_strings)
    print JSON.parse(make_kubernetes_api_get_request(uri + query).body)
    return JSON.parse(make_kubernetes_api_get_request(uri + query).body)["items"]
end

map = {}
map['app'] = 'nginx'
#print pod_names_with_labels('test-rbac', map)
print scale('nginx', 4, 'async')