=begin
1. Identify an existing pods (targets to preserve)
2. Change label to be outside of deployment -> new pods will spawn in their place [ possibly all existing pods also ]
3. Downscale the deployment to max(0, desired_scale - exempted_pods), then upscale to desired_scale
4. Add back these pods to deployment

 Case 1: Number of exempted pods <= New scale count -> Tested
 -> Some 'good' pods will die and 'exempted' pods take their place
 -> Number of pods in system ==  New scale count

 Case 2: Number of exempted pods > New scale count -> Tested
 -> Other pods will die off and some of these pods will in GS, rest will serve traffic
 -> All pods in the system are 'exempted'
 -> Number of pods in system ==  Number of exempted pods
 -> Number of exempted pods - New scale count are extra in the system
=end

require 'json'

def namespace
  'alphabet --context gke_gcp-controlplane_us-east1-b_qds-k8s-dev-private'
end

def scale(deployment:, replicas:)
  `kubectl scale deployment #{deployment} -n #{namespace} --replicas #{replicas}`
end

def wait(deployment)
  `kubectl rollout status deployment #{deployment} -n #{namespace}`
end

def exempted_pods
  %w[client-59cf444995-vrxl2]
end

def computed_exempted_pods(deployment)
  exempted_pods
end

def detach_pods(pods:, deployment_label_key:, deployment_label_value:)
  deployment_label_value += 'graceful-shutdown-exempted'
  pods.each { |pod| patch_pod(pod: pod, deployment_label_key: deployment_label_key, deployment_label_value: deployment_label_value) }
end

def re_attach_pods(pods:, deployment_label_key:, deployment_label_value:)
  pods.each { |pod| patch_pod(pod: pod, deployment_label_key: deployment_label_key, deployment_label_value: deployment_label_value) }
end

def metadata(pod:)
  JSON.parse(`kubectl get pods -n #{namespace} -o json #{pod}`)['metadata']
end

def patched_metadata(pod:, deployment_label_key:, deployment_label_value:)
  pod_metadata = metadata(pod: pod)
  labels = pod_metadata['labels'] || {}
  labels[deployment_label_key] = deployment_label_value
  pod_metadata['labels'] = labels

  pod_metadata
end

def patch_json(pod:, deployment_label_key:, deployment_label_value:)
  { metadata: patched_metadata(pod: pod, deployment_label_key: deployment_label_key, deployment_label_value: deployment_label_value) }
end

def patch_pod(pod:, deployment_label_key:, deployment_label_value:)
  `kubectl -n #{namespace} patch pod #{pod} --patch '#{patch_json(pod: pod, deployment_label_key: deployment_label_key, deployment_label_value: deployment_label_value).to_json}'`
end

def downscale(target_scale:, deployment_name:, deployment_label_key:, deployment_label_value:)
  exempted_pods = computed_exempted_pods(deployment_name)
  number_of_exempted_pods = exempted_pods.length
  puts "detaching #{exempted_pods}"
  detach_pods(pods: exempted_pods, deployment_label_key: deployment_label_key, deployment_label_value: deployment_label_value)
  # sleep 3
  # puts "scaling to #{[target_scale - number_of_exempted_pods, 0].max}"
  scale(deployment: deployment_name, replicas: [target_scale - number_of_exempted_pods, 0].max)
  # sleep 3
  # puts "scaling to #{target_scale}"
  scale(deployment: deployment_name, replicas: target_scale)
  # puts "attaching to #{exempted_pods}"
  # sleep 3
  re_attach_pods(pods: exempted_pods, deployment_label_key: deployment_label_key, deployment_label_value: deployment_label_value)
  wait(deployment_name)
end

def upscale(target_scale:, deployment_name:)
  scale(deployment: deployment_name, replicas: target_scale)
end

def main
  current_scale = 3
  target_scale = 2
  deployment_label_key = 'name'
  deployment_label_value = 'client'
  deployment_name = 'client'
  # Downscaling
  upscale(target_scale: current_scale, deployment_name: deployment_name)
  wait(deployment_name)
  downscale(target_scale: target_scale, deployment_name: deployment_name, deployment_label_key: deployment_label_key, deployment_label_value: deployment_label_value)
end

main
