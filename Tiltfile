docker_build('271036156099.dkr.ecr.us-west-1.amazonaws.com/api-server', '', dockerfile="api-server.dockerfile")
docker_build('271036156099.dkr.ecr.us-west-1.amazonaws.com/capi-api', '', dockerfile="capi-api.dockerfile")
docker_build('271036156099.dkr.ecr.us-west-1.amazonaws.com/aerostation-capi-controller', '', dockerfile="Dockerfile")
docker_build('271036156099.dkr.ecr.us-west-1.amazonaws.com/user-service', '', dockerfile="user-service.dockerfile")

k8s_yaml(kustomize('./config/crd'))
k8s_yaml('./app/controller.yaml')
k8s_yaml('./api-server/app/deploy.yaml')
k8s_yaml('./capi-api/app/deploy.yaml')
k8s_yaml('./user-service/app/deploy.yaml')

k8s_resource('api-server', port_forwards='9000')
k8s_resource('user-service', port_forwards='8081')

allow_k8s_contexts('Administrator@fleet-controller-manager.us-west-1.eksctl.io')