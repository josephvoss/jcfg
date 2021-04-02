{
  local core = import 'util/core.jsonnet',
  // Build systemd section content.
  // Args:
  //   section  Name of the section to create
  //   settings Map of the key values settings to set for section
  local systemd_section_content(section, settings) =
    std.join(
      '\n',
      ['[' + section + ']'] +
      ['%s=%s' % [key, settings[key]] for key in std.objectFields(settings)]
    ) + '\n',

  // Create systemd unit file
  // params:
  //   path           Path to file to install. Defaults to /etc/systemd/system/name.unit
  //   unit_values    Map of unit keys and values
  //   install_values Map of install keys and values
  //   File (Passed to file resource directly)
  //
  UnitFile(name, params):: {
    local content_string =
      systemd_section_content('Unit', params.unit_values) +
      systemd_section_content('Install', params.install_values),
    local f_params = {
      name: name,
      path: '/etc/systemd/system/' + name + '.unit',
      content: { type: 'string', string: content_string },
    } + params.file_params,
    output: core.File('', f_params),
  },
  ServiceFile(name, params):: {
    local content_string =
      systemd_section_content('Unit', params.unit_values) +
      systemd_section_content('Install', params.install_values) +
      systemd_section_content('Service', params.service_values),
    local f_params = {
      name: name,
      path: '/etc/systemd/system/' + name + '.service',
      content: { type: 'string', string: content_string },
    } + params.file_params,
    output: core.File('', f_params),
  },
  // Defines state of unit on the system
  // params:
  //   active
  //   enable
  //   ordering
  //
  Unit(name, params):: {
    local set_enable = params.enable,
    // Check enabled - fail, set enabled
    local check_enabled = {
      name: 'Check %s enabled state' % name,
      path: '/bin/systemctl',
      args: ['is-enabled', name],
      exitcode: if set_enable == true then 0 else 1,
      failOk: true,
    },
    local set_enabled = {
      name: 'Set %s enabled state' % name,
      path: '/bin/systemctl',
      args: [if set_enable == true then 'enable' else 'disable', name],
      failOk: false,
      ordering: { afterFail: 'Check %s enabled state' % name },
    },
    local param_active = params.active,
    // Check active - fail, set active
    local check_active = {
      name: 'Check %s active state' % name,
      path: '/bin/systemctl',
      args: ['is-active', name],
      exitcode: if param_active == true then 0 else 1,
      failOk: true,
      ordering: { afterOk: params.ordering },
    },
    local set_active = {
      name: 'Set %s active state' % name,
      path: '/bin/systemctl',
      args: [if param_active == true then 'start' else 'stop', name],
      failOk: false,
      ordering: { afterFail: 'Check %s active state' % name },
    },
    output: [
      core.Exec('', check_enabled),
      core.Exec('', set_enabled),
      core.Exec('', check_active),
      core.Exec('', set_active),
    ],
  },
}
