require_relative 'hipaa_helper'
require 'rest-client'
require 'json'

def service_exists?(service)
  begin
    r = KubernetesApiHelper.make_kubernetes_api_call(
    method: :get,
    uri: "/api/v1/namespaces/#{ENV['NAMESPACE']}/services/#{service}",
    )
    return r.code == 200
  rescue
    return false
  end
end

def unique_identifier_labels
  {
    "pod_ip" => ENV['POD_IP']
  }
end

def create_client_service
  delete_service(client_service_name) if service_exists?(client_service_name)
  data = current_pod_data
  selectors = data['metadata']['labels']
  unique_identifier_labels.each do |k,v|
    selectors[k] = v
  end
  name = client_service_name
  payload = {
    "kind"=>"Service", 
    "apiVersion"=>"v1", 
    "metadata"=> {
      "name"=>"#{name}",
      "namespace"=>"#{ENV['NAMESPACE']}",
      "labels" => { 
        "group"=>"services",
        "target"=>"conduit-#{ENV['CONDUIT'].gsub(".","-")}"
        }
      }, 
      "spec"=> {
        "selector"=> selectors, 
        "type"=>"ClusterIP", 
        "ports"=>[
          {
            "name"=>"http-#{name}-jeeves",
            "protocol"=>"TCP", "port"=>80
          }, 
          {
            "name"=>"http-#{name}-consul1", 
            "protocol"=>"TCP", "port"=>8500
          },
          {
            "name"=>"http-#{name}-consul2",
            "protocol"=>"TCP", "port"=>8301
            }
          ]
        }
      }.to_json
  
  headers = {'Content-Type'=> 'application/json', 'Accept'=> '*/*'} 
    KubernetesApiHelper.make_kubernetes_api_call(
    method: :post,
    uri: "/api/v1/namespaces/#{ENV['NAMESPACE']}/services",
    headers: headers,
    payload: payload
  )
end

def add_pod_identifer_labels
  data = current_pod_data
  unique_identifier_labels.each do |k,v|
    data["metadata"]["labels"][k] = v
  end
  headers = {'Content-Type'=> 'application/json-patch+json', 'Accept'=> 'application/json'} 
  payload = [{"op" => "add", "path" => "/metadata", "value" => data["metadata"]}].to_json.to_s
  KubernetesApiHelper.make_kubernetes_api_call(
    method: :patch,
    uri: "/api/v1/namespaces/#{ENV['NAMESPACE']}/pods/#{ENV['POD_NAME']}",
    headers: headers,
    payload: payload
  )
end

add_pod_identifer_labels
create_client_service
