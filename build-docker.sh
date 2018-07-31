#!/bin/bash

tag=$1

if [ -z $tag ]; then
  echo "please provide a tag arg"
  exit 1
fi

major=$(echo $tag | awk -F. '{print $1}')
minor=$(echo $tag | awk -F. '{print $2}')
# patch=$(echo $tag | awk -F. '{print $3}')

tf_ver=$(grep TERRAFORM_VERSION Dockerfile | head -n 1 | awk '{print $3}')

docker build -t jmccann/drone-terraform:latest .

set -x
docker tag jmccann/drone-terraform:latest jmccann/drone-terraform:${major}
docker tag jmccann/drone-terraform:latest jmccann/drone-terraform:${major}.${minor}
docker tag jmccann/drone-terraform:latest jmccann/drone-terraform:${major}.${minor}-${tf_ver}

docker push jmccann/drone-terraform:latest
docker push jmccann/drone-terraform:${major}
docker push jmccann/drone-terraform:${major}.${minor}
docker push jmccann/drone-terraform:${major}.${minor}-${tf_ver}
set +x