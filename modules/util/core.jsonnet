{
  // Params are
  // (Mandatory)
  // name
  // path
  // content
  // (Optional)
  // ensure
  // owner
  // group
  // mode
  // ordering
  File(name, params):: {
    local f_params = {
      name: name,
      ordering: {},
      ensure: 'present',
      userid: { owner: 'root', group: 'root' },
      mode: '0644',
    } + params,
    api: 'v1',
    kind: 'File',
    metadata: {
      name: f_params.name,
      ordering: f_params.ordering,
    },
    spec: {
      ensure: f_params.ensure,
      userid: f_params.userid,
      mode: f_params.mode,
      path: f_params.path,
      content: f_params.content,
    },
  },
  Exec(name, params):: {
    local e_params = {
      name: name,
      path: name,
      ordering: {},
      userid: { owner: 'root', group: 'root' },
      args: [],
      env: [],
      dir: '/',
      timeout: '',
      exitcode: 0,
      failOk: false,
    } + params,
    api: 'v1',
    kind: 'Exec',
    metadata: {
      name: e_params.name,
      ordering: e_params.ordering,
    },
    spec: {
      path: e_params.path,
      args: e_params.args,
      env: e_params.env,
      dir: e_params.dir,
      userid: e_params.userid,
      timeout: e_params.timeout,
      exitcode: e_params.exitcode,
      failOk: e_params.failOk,
    },
  },
}
