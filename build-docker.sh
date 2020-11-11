#!/bin/bash

set -e

tag=$1

if [ -z $tag ]; then
  echo "please provide a tag arg"
  exit 1
fi

major=$(echo $tag | awk -F. '{print $1}')
minor=$(echo $tag | awk -F. '{print $2}')

tf_ver="0.13.5"

echo "Confirm building images for:"
echo "  MAJOR: ${major}"
echo "  MINOR: ${minor}"
echo "  TF_VERSION: ${tf_ver}"

read -p "Proceed? [Y/N] " ans

if [[ "$ans" != "Y" && "$ans" != "y" ]]; then
  echo "Cancelling"
  exit 0
fi

set -x
docker build -t jmccann/drone-terraform:latest --build-arg terraform_version=${tf_ver} .

docker tag jmccann/drone-terraform:latest jmccann/drone-terraform:${major}
docker tag jmccann/drone-terraform:latest jmccann/drone-terraform:${major}.${minor}
docker tag jmccann/drone-terraform:latest jmccann/drone-terraform:${major}.${minor}-${tf_ver}

docker push jmccann/drone-terraform:latest
docker push jmccann/drone-terraform:${major}
docker push jmccann/drone-terraform:${major}.${minor}
docker push jmccann/drone-terraform:${major}.${minor}-${tf_ver}
set +x
