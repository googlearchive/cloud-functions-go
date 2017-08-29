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

Vagrant.configure("2") do |config|
  config.vm.provider :virtualbox do |_, override|
    override.vm.box = "debian/stretch64"
    override.vm.network "forwarded_port", guest: 8080, host: 8080, auto_correct: true
    override.vm.synced_folder ".", "/vagrant", type: "virtualbox"
  end
  config.vm.provision :shell, path: "provision.sh"
end
