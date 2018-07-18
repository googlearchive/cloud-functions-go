#!/usr/bin/env bash
# Copyright 2017 Google Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -euo pipefail

apt-get update
apt-get dist-upgrade -y
apt-get install -y zip curl build-essential

curl -sL https://deb.nodesource.com/setup_6.x | bash -
apt-get install -y nodejs node-gyp npm

mkdir /tmp/go
cd /tmp/go
wget https://godeb.s3.amazonaws.com/godeb-amd64.tar.gz
tar -xvzf godeb-amd64.tar.gz
./godeb install
cd
rm -rf /tmp/go

go get -d -u google.golang.org/api/pubsub/v1 google.golang.org/api/storage/v1
