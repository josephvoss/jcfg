{
  local core = import 'util/core.jsonnet',
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

  Yum_repo_file(name, params):: {
    local f_params = {
      name: name,
      content: { type: 'string', string: yum_repo_content(repo_params) },
    } + params.file_params,
    local repo_params = {
      name: name,
      baseurl: 'https://mirror.example.com/' + (
        if std.objectHas(params, 'snapshot') then
          'snapshot/' + params.snapshot + '/' + name
        else
          'custom/' + name + '/'
      ) + (
        if std.objectHas(params, 'os_arch') then
          params.os_arch
        else
          'el-7-x86_64/'
      ),
    } + params.repo_params,
    output: core.File('', f_params),
  },
}
