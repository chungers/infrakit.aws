#!/bin/bash

# initializeManager
set -o errexit
set -o nounset
set -o xtrace

# See http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/device_naming.html why device naming is tricky, and likely
# coupled to the AMI (host OS) used.
EBS_DEVICE=/dev/xvdf

# Determine whether the EBS volume needs to be formatted.
if [ "$(file -sL $EBS_DEVICE)" = "$EBS_DEVICE: data" ]
then
  echo 'Formatting EBS volume device'
  mkfs -t ext4 $EBS_DEVICE
fi

systemctl stop docker
rm -rf /var/lib/docker

mkdir -p /var/lib/docker
echo "$EBS_DEVICE /var/lib/docker ext4 defaults,nofail 0 2" >> /etc/fstab
mount -a
systemctl start docker

# startInfrakit

plugins=/infrakit/plugins
configs=/infrakit/configs
discovery="-e INFRAKIT_PLUGINS_DIR=$plugins -v $plugins:$plugins"
local_store="-v /infrakit/:/infrakit/"
docker_client="-v /var/run/docker.sock:/var/run/docker.sock"
run_plugin="docker run -d --restart always $discovery"
image=wfarner/infrakit-demo-plugins
manager=chungers/demo

mkdir -p $configs
mkdir -p $plugins

docker pull $image
docker pull $manager
$run_plugin --name flavor-combo $image infrakit-flavor-combo --log 5
$run_plugin --name flavor-swarm $docker_client $image infrakit-flavor-swarm --log 5 --name flavor-swarm
$run_plugin --name flavor-vanilla $image infrakit-flavor-vanilla --log 5
$run_plugin --name group-stateless $image infrakit-group-default --name group-stateless --log 5
$run_plugin --name instance-aws $image infrakit-instance-aws --log 5
$run_plugin --name manager $docker_client $manager infrakit-manager swarm --proxy-for-group group-stateless --name group --log 5

echo "alias infrakit='docker run --rm $discovery $local_store $docker_client $manager infrakit'" >> /home/ubuntu/.bashrc
echo "alias infrakit='docker run --rm $discovery $local_store $docker_client $manager infrakit'" >> /root/.bashrc
