require_relative 'hipaa_helper'
require 'rest-client'
require 'json'
require 'timeout'
 
 def add_back_pod_to_endpoint
   headers = {'Content-Type'=> 'application/json-patch+json', 'Accept'=> 'application/json'} 
   payload = [{"op" => "add", "path" => "/subsets", "value" => subsets_payload}].to_json.to_s
   KubernetesApiHelper.make_kubernetes_api_call(
     method: :patch,
     uri: "/api/v1/namespaces/#{ENV['NAMESPACE']}/endpoints/#{client_service_name}",
     headers: headers,
     payload: payload
   )
 end
 
 def subsets_payload
   data = endpoint_data
   p = current_pod_data
   addresses_data = {
     "ip" => ENV['POD_IP'],
     "nodeName" => p['spec']['nodeName'],
     "targetRef" => {
       "kind" => "Pod",
       "name" => ENV['POD_NAME'],
       "namespace" =>   ENV['NAMESPACE'],
       "resourceVersion" => p['metadata']['resourceVersion'],
       "uid" =>  p['metadata']['uid'],
     }
   }
   
   if !data["subsets"].nil?
     data["subsets"][0]["addresses"].push(addresses_data)
   else
     data["subsets"] = [ {"addresses" => [addresses_data] } ]
   end
  return data["subsets"]
 end
 
 def wait_till_pod_out_of_service
   Timeout::timeout(300) do
     while pod_in_service?(client_service_name, ENV['POD_IP'])
       sleep 1
     end
   end
 end

wait_till_pod_out_of_service
add_back_pod_to_endpoint

