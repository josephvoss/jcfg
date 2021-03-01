local package = import 'yum/package.jsonnet';

local pkg_params = {
};

std.flattenArrays([
  package.Package(
    'kernel-devel', pkg_params
  ).output,
])
