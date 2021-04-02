// Need repos, packages, users, services

// Users
local users = import 'modules/user/user.jsonnet';
local user_hash = import 'secrets/decrypted/user_hashes.json';
local jvoss_user_hash = user_hash.jvoss_user;
local root_user_hash = user_hash.root_user;
local local_jvoss_user = users.User('local-jvoss', { password: jvoss_user_hash }).output;
local root_user = users.User('root', { password: root_user_hash }).output;

// Services
local unit = import 'modules/systemd/unit.jsonnet';
local after_ok_ssh_server_pkg = { afterOk: ['Exec::Install package openssh-server'] };
local sshd_unit = unit.Unit('sshd.service', {
  enable_ordering: after_ok_ssh_server_pkg,
  active_ordering: after_ok_ssh_server_pkg,
}).output;

// Repos
local yumrepo = import 'modules/yum/yumrepo.jsonnet';
local gpg_enabled = {
  gpgcheck: 1,
  enabled: 1,
};
local rhel_repo = gpg_enabled {
  gpgkey: 'file:///etc/pki/rpm-gpg/RPM-GPG-KEY-redhat-release',
};
local rhel_upstream_base = 'https://mirror.example.com/rhel/7/x86_64/';
local repo_hash = {
  install: rhel_repo { baseurl: 'https://mirror.example.com/rhel7-x86_64/' },
  os: rhel_repo { baseurl: rhel_upstream_base + 'os' },
  supplementary: rhel_repo { baseurl: rhel_upstream_base + 'supplementary' },
  optional: rhel_repo { baseurl: rhel_upstream_base + 'optional' },
  extras: rhel_repo {
    baseurl: 'https://mirror.example.com/rhel/7Server/x86_64/extras',
  },
  epel: {
    baseurl: 'https://mirror.example.com/el/7/x86_64/epel',
    gpgkey: 'https://mirror.example.com/keys/RPM-GPG-KEY-EPEL-7',
  },
};

// Packages
local pkg = import 'modules/yum/package.jsonnet';
local sshd_pkg = pkg.Package('openssh-server', {
  ordering: { afterOk: ['File::Yum Repo file for install'] },
}).output;

sshd_unit + local_jvoss_user + root_user + [
  yumrepo.Yum_repo_file(k, { repo_params: repo_hash[k] }).output
  for k in std.objectFieldsAll(repo_hash)
] + sshd_pkg
