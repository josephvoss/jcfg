{
  local core = import 'util/core.jsonnet',
  Package(name, params):: {
    // couple of resources here to be chained
    // check if package installed. If err, run follow to install
    // If that bombs, fail
    local input_ordering = if std.objectHas(params, 'ordering') then params.ordering else {},
    local package_name = name,
    local yum_args =
      if std.objectHas(params, 'yum_args') then
        params.yum_args
      else
        ['-d', '1', '-y'],
    local check_install = {
      name: 'Check package %s installed' % name,
      ordering: input_ordering,
      path: '/usr/bin/rpm',
      args: ['-q', name],
      exitcode: 0,
      failOk: true,
    },
    local do_install = {
      name: 'Install package %s' % name,
      ordering: { after: ['Exec::Check package %s installed' % name] },
      path: '/usr/bin/yum',
      args: ['install'] + yum_args + [name],
      exitcode: 0,
    },
    output: [
      core.Exec('', check_install),
      core.Exec('', do_install),
    ],
  },
}
