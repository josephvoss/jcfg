local resource_name = 'temp_repo';
local enabled = 1;
local owner = 'root';
local group = 'root';
local mode = '0644';
local base_dir = '/etc/yum.repos.d/';

local ensure = 'present';
local ordering = {};

local repo_name = 'temp_repo';
local baseurl = 'http://mirrors.fedoraproject.org/mirrorlist?repo=fedora-34&arch=x86_64'
local gpgcheck = 0;
local gpgkey = '';
local enabled = 1;

local path = base_dir + '/' + repo_name + '.repo';

local yumrepo = import 'yum/yumrepo.jsonnet';

local f_params = {
  // resource_name: resource_name,
  ensure: ensure,
  userid: { owner: owner, group: group },
  mode: mode,
  path: path,
  base_dir: base_dir,
  ordering: ordering,
};

local yum_repo_params = {
  //  name: repo_name,
  // baseurl: baseurl,
  gpgcheck: gpgcheck,
  gpgkey: gpgkey,
  enabled: enabled,
};

[
  yumrepo.Yum_repo_file(
    'temp_repo_1', { file_params: f_params, repo_params: yum_repo_params },
  ).output,
  yumrepo.Yum_repo_file(
    'temp_repo_2', { snapshot: 2020226, file_params: f_params, repo_params: yum_repo_params },
  ).output,
]
