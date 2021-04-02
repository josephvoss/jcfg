{
  local core = import 'modules/util/core.jsonnet',

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
      args: ['-u', p.name],
      exitcode: 0,
      failOk: false,
      ordering: p.ordering,
    },
    // Shared args between useradd and usermod
    local shared_user_args = std.prune(std.flattenArrays([
      if p.password != '' then ['--password', p.password] else [],
      if p.shell != '' then ['--shell', p.shell] else [],
      if p.uid != '' then ['--uid', p.uid] else [],
      if p.gid != '' then ['--gid', p.gid]
      else if p.group != '' then ['-g', p.group] else [],
      if p.home != '' then ['-d', p.home] else [],
    ])),
    local useradd_args = std.prune(std.flattenArrays([shared_user_args, [
      if p.create_home then '--create-home' else '--no-create-home',
      // TODO: Logic here might be off, not sure of -U vs -g conflicts
      if p.create_group then '--user-group' else '--no-user-group',
      if p.system then '--system',
      p.name,
    ]])),
    local usermod_args = std.prune(std.flattenArrays([shared_user_args, [p.name]])),
    local create_user = {
      name: 'Create user %s' % p.name,
      path: '/usr/sbin/useradd',
      args: useradd_args,
      exitcode: 0,
      failOk: false,
      ordering: { afterFail: ['Exec::Check user %s exists' % p.name] },
    },
    // Check user attributes - fail, mod
    // Well, we can't actually easily check these params...mod every time I guess :/
    local mod_user = {
      name: 'Modify user %s' % p.name,
      path: '/usr/sbin/usermod',
      args: usermod_args,
      exitcode: 0,
      failOk: false,
      ordering: { afterOk: ['Exec::Check user %s exists' % p.name] },
    },
    output: [
      core.Exec('', check_exists),
      core.Exec('', create_user),
      core.Exec('', mod_user),
    ],
  },
}
