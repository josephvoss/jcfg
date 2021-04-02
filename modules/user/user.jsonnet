{
  local core = import 'util/core.jsonnet',

  // Create local user
  // params:
  //   name User name
  //   password Password hash to user
  //   uid UID num to set
  //   group User's login group, must already exist
  //   gid Login group as num, must already exist
  //   home Home directory location
  //   create_home Make the home dir
  //   create_group Make the home dir
  //   system Is this a system account?
  //   ordering Schedules initial check resource
  //
  // Graph:
  //   if Check user exist == ok then Modify user else Create user
  //
  User(name, params):: {
    // Set defaults
    local p = {
      name: name,
      passsword: '',
      shell: '',
      uid: '',
      group: '',
      gid: '',
      home: '',
      create_home: true,
      create_group: true,
      system: false,
      ordering: {},
    } + params,
    // Check user exists - fail, create
    // id returns 0 if user exists, 1 if not
    local check_exists = {
      name: 'Check user %s exists' % p.name,
      path: '/usr/bin/id',
      args: ['-u ', p.name],
      exitcode: 0,
      failOk: true,
      ordering: p.ordering,
    },
    // Shared args between useradd and usermod
    local shared_user_args = [
      if p.password != '' then '--password %s' % p.password,
      if p.shell != '' then '--shell %s' % p.shell,
      if p.uid != '' then '--uid %s' % p.uid,
      if p.gid != '' then '--gid %s' % p.gid else if p.group != '' then '-g %s' % p.group,
      if p.home != '' then '-d %s' % p.home,
    ],
    local useradd_args = shared_user_args + [
      if p.create_home then '--create-home' else '--no-create-home',
      // TODO: Logic here might be off, not sure of -U vs -g conflicts
      if p.create_group then '--user-group' else '--no-user-group',
      if p.system then '--system',
    ],
    local usermod_args = shared_user_args,
    local create_user = {
      name: 'Create user %s' % p.name,
      path: '/usr/sbin/useradd',
      args: useradd_args,
      exitcode: 0,
      failOk: false,
      ordering: { afterFail: 'Check user %s exists' % p.name },
    },
    // Check user attributes - fail, mod
    // Well, we can't actually easily check these params...mod every time I guess :/
    local mod_user = {
      name: 'Modify user %s' % p.name,
      path: '/usr/sbin/usermod',
      args: usermod_args,
      exitcode: 0,
      failOk: false,
      ordering: { afterOk: 'Check user %s exists' % p.name },
    },
    output: [
      core.Exec('', check_exists),
      core.Exec('', create_user),
      core.Exec('', mod_user),
    ],
  },
}
