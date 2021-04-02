{
  local core = import 'modules/util/core.jsonnet',
  local yum_repo_content(
    //    name,
    //    baseurl,
    //    enabled=1,
    //    gpgcheck=0,
    //    gpgkey='',
    yrcp
  ) =
    // Join array inputs w/ newline to build content string. Add newline to the
    // end of string. If not set, don't include
    std.join(
      '\n',
      ['[' + yrcp.name + ']'] +
      [
        param + ' = ' + yrcp[param]
        for param in std.objectFields(yrcp)
        if yrcp[param] != ''
      ],
    ) + '\n',

  // Params:
  //    baseurl
  //    gpg
  Yum_repo_file(name, params):: {
    local p = {
      name: name,
      snapshot: '',
      os_arch: '',
      // Params directly to file resource
      file_params: {},
      // Repo params, essentially key/values for repo config
      repo_params: {},
    } + params,
    // Defaults for repo params, try to set baseurl
    local repo_params = {
      name: p.name,
      baseurl: 'https://mirror.example.com/' + (
        if p.snapshot != '' then
          'snapshot/' + p.snapshot + '/' + p.name
        else
          'custom/' + name + '/'
      ) + (
        if std.objectHas(params, 'os_arch') then
          params.os_arch
        else
          'el7-x86_64/'
      ),
    } + p.repo_params,
    // File resource to make, over write settings w/ file_param.s
    // Content set by helper function w/ repo_params as input
    local f_params = {
      name: 'Yum Repo file for %s' % p.name,
      ensure: 'present',
      userid: { owner: 'root', group: 'root' },
      mode: '0644',
      path: '/etc/yum.repos.d/%s.repo' % p.name,
      content: { type: 'string', string: yum_repo_content(repo_params) },
      ordering: {},
    } + p.file_params,
    output: core.File('', f_params),
  },
}
