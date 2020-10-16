#!/usr/bin/env ruby
require 'rest-client'

module KubernetesApiHelper
  def self.make_kubernetes_api_call(method: :get, uri: '', headers: {}, payload: {})
    base_url = "https://kubernetes.default.svc.cluster.local"
    url = base_url + uri
    ca_cert_file = File.join('/var', 'run', 'secrets', 'kubernetes.io', 'serviceaccount', 'ca.crt')
    auth_token = File.read(File.join('/var', 'run', 'secrets', 'kubernetes.io', 'serviceaccount', 'token'))
    headers['Authorization'] = "Bearer #{auth_token}"
    RestClient::Request.execute(method: method, url: url, ssl_ca_file: ca_cert_file, headers: headers, payload: payload)
  end
end


def client_service_name
  "client-#{ENV['CONDUIT'].gsub(".","-")}-#{ENV['POD_IP'].gsub('.','-')}"
end

def endpoint_data(svc=client_service_name)
  JSON.parse(
    KubernetesApiHelper.make_kubernetes_api_call(
    method: :get,
    uri: "/api/v1/namespaces/#{ENV['NAMESPACE']}/endpoints/#{svc}",
  )
)
end

def pod_data(pod)
 JSON.parse(KubernetesApiHelper.make_kubernetes_api_call(
    method: :get,
    uri: "/api/v1/namespaces/#{ENV['NAMESPACE']}/pods/#{pod}",
  )
)
end

def delete_service(svc)
    KubernetesApiHelper.make_kubernetes_api_call(
    method: :delete,
    uri: "/api/v1/namespaces/#{ENV['NAMESPACE']}/services/#{svc}",
  )
end

def pod_in_service?(service, pod_ip)
 r= KubernetesApiHelper.make_kubernetes_api_call(
    method: :get,
    uri: "/api/v1/namespaces/#{ENV['NAMESPACE']}/endpoints/#{service}",
  )

  return JSON.parse(r)['subsets'][0]['addresses'].map { |k| k["ip"] }.include?(pod_ip) rescue false
end

def current_pod_data
  pod_data(ENV["POD_NAME"])
end