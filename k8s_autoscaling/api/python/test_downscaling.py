import random
import time
from k8s_apis_helper import scale, deployment_pod_labels, pod_names_with_labels, check_pods_exist
from downscaling import exempt_pods_from_downscale, downscale


def test_downscale_within_exempt_count(deployment, start_scale, target_scale, exempt_count, namespace):
    if target_scale > start_scale:
        print "Test is for downscaling, please provide target_scale < start_scale"
        exit(1)
    if exempt_count > target_scale:
        print "Test is for downscaling within exempt_count, please provide exempt_count < target_scale"
        exit(1)
    deployment_labels = deployment_pod_labels(
        deployment=deployment, namespace=namespace)
    scale(deployment=deployment, replicas=0, namespace=namespace)
    while pod_names_with_labels(namespace, deployment_labels):
        time.sleep(1)
    # Upscale to start count
    scale(deployment=deployment, replicas=start_scale, namespace=namespace)
    # Wait for completion (shoud use rollout if possible)
    time.sleep(int(start_scale / 10 + 1) * 20)

    #pods_in_deployment = pod_names_with_labels(namespace, deployment_labels)
    exempted_pods = random.sample(pod_names_with_labels(
        namespace, deployment_labels), k=exempt_count)
    # Add labels to some pods to prevent them from being removed during downscaling.
    # This label should be added by pod itself as a cron
    exempt_pods_from_downscale(
        exempted_pods=exempted_pods, namespace=namespace)
    downscale_start_time = time.time()
    downscale(target_scale=target_scale, deployment_name=deployment,
              deployment_labels=deployment_labels, namespace=namespace)
    print time.time() - downscale_start_time
    # Ensure pod count is correct after downscale
    # Cannot be part of downscale method as it is generic and GS can be long running.
    while len(pod_names_with_labels(namespace, deployment_labels)) > target_scale:
        time.sleep(1)
    new_pods = pod_names_with_labels(namespace, deployment_labels)
    # Ensure no pods marked as exempted_pods got deleted
    check_pods_exist(expected_pods=exempted_pods, present_pods=new_pods)
    # Teardown
    scale(deployment=deployment, replicas=0, namespace=namespace)


DEPLOYMENT = "nginx"
NAMESPACE = "default"
try:
    while True:
        START_SCALE, TARGET_SCALE, EXEMPT_COUNT = map(int, raw_input().split())
        # start_scale,target_scale,exempt_count = int(sys.argv[1]),int(sys.argv[2]), int(sys.argv[3])
        test_downscale_within_exempt_count(deployment=DEPLOYMENT, start_scale=START_SCALE,
                                           target_scale=TARGET_SCALE, exempt_count=EXEMPT_COUNT, namespace=NAMESPACE)
        scale(deployment=DEPLOYMENT, replicas=0, namespace=NAMESPACE)
except:
    pass
