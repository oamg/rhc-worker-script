# Specfile based and updated from
# https://github.com/theforeman/foreman-packaging/blob/rpm/develop/packages/client/foreman_ygg_worker/foreman_ygg_worker.spec

%define debug_package %{nil}

# Flags for building the package
%global buildflags -buildmode pie -compiler gc -a -v -x
%global goldflags %{expand:-linkmode=external -compressdwarf=false -B 0x$(head -c20 /dev/urandom|od -An -tx1|tr -d ' \\n') -extldflags '%__global_ldflags'}

# Package constants
%global repo_orgname oamg
%global repo_name rhc-worker-script
%global binary_name rhc-script-worker
%global rhc_libexecdir %{_libexecdir}/rhc
%{!?_root_sysconfdir:%global _root_sysconfdir %{_sysconfdir}}
%global rhc_worker_conf_dir %{_root_sysconfdir}/rhc/workers

# Go toolset configuration
%global go_toolset_version 1.19

# EL7 doesn't define go_arches (it is available in go-srpm-macros which is EL8+)
%if !%{defined go_arches}
%define go_arches x86_64
%endif

%global use_go_toolset_1_19 0%{?rhel} == 7 && !%{defined centos}

Name:           %{repo_name}
Version:        0.8
Release:        1%{?dist}
Summary:        Worker executing scripts on hosts managed by Red Hat Insights

License:        GPLv3+
URL:            https://github.com/%{repo_orgname}/%{repo_name}
Source0:        %{url}/releases/download/v%{version}/%{name}-%{version}.tar.gz
ExclusiveArch:  %{go_arches}

BuildRequires: git
%if %{use_go_toolset_1_19}
BuildRequires:  go-toolset-%{go_toolset_version}-golang
%else
BuildRequires:  golang
%endif
Requires:       rhc

%description
Remote Host Configuration (rhc) worker for executing scripts on hosts
managed by Red Hat Insights.

%prep
%setup -q

%build
export CGO_CPPFLAGS="-D_FORTIFY_SOURCE=2 -fstack-protector-all"
export BUILDFLAGS="%{buildflags}"
export LDFLAGS="%{goldflags}"

%if %{use_go_toolset_1_19}
scl enable go-toolset-%{go_toolset_version} -- make build
%else
make build
%endif

%install
# Create a temporary directory /var/lib/rhc-worker-script - used mainly for storing temporary files
install -d %{buildroot}%{_sharedstatedir}/%{binary_name}/

install -D -m 755 build/%{binary_name} %{buildroot}%{rhc_libexecdir}/%{binary_name}
install -D -d -m 755 %{buildroot}%{rhc_worker_conf_dir}

cat <<EOF >%{buildroot}%{rhc_worker_conf_dir}/rhc-worker-script.yml
# Recipient directive to register with dispatcher
directive: "%{name}"

# Whether to verify incoming yaml files
verify_yaml: true

# Temporary directory in which the temporary script will be placed and executed.
temporary_worker_directory: "/var/lib/rhc-worker-script"

# Pass environment variables to the script being executed
# env:
  # environment variables to be set for the script
  # FOO: "some-string-value"
  # BAR: "other-string-value"

# Log level that will be sent to the script
script_log_level: "info"
EOF

%files
%{rhc_libexecdir}/%{binary_name}
%license LICENSE
%doc README.md
%config %{rhc_worker_conf_dir}/rhc-worker-script.yml

%changelog

* Wed Apr 24 2024 Rodolfo Olivieri <rolivier@redhat.com> 0.8-1
- Pass log level to executed script for more granular logging possibility
- Bump golang.org/x/net from 0.17.0 to 0.23.0
- Refactor specfile for building the worker package

* Wed Apr 10 2024 Rodolfo Olivieri <rolivier@redhat.com> 0.7-1
- Load env vars from worker config file into script execution env

* Wed Feb 28 2024 Rodolfo Olivieri <rolivier@redhat.com> 0.6-1
- Fix grpc to newest v1.59.x version
- Remove insights_core_gpg_check from worker config
- When script fails with exit code 1 we want to see the reason in logs

* Mon Oct 16 2023 Rodolfo Olivieri <rolivier@redhat.com> 0.5-1
- Rebuild against newer golang which addresses CVE-2023-39325 and CVE-2023-44487
- Fix OpenScanHub defects related to runtime code
- Update specfile to include default config file
- Improve logging when config file can't be used and default values are used instead
- Move the logging init to be before anything else
- Improve logging when insights_core_gpg_check is disabled

* Thu Aug 10 2023 Rodolfo Olivieri <rolivier@redhat.com> 0.4-1
- Update specfile binary name generation
- Add couple more unit tests for util.go

* Thu Aug 10 2023 Rodolfo Olivieri <rolivier@redhat.com> 0.3-1
- Parse minimal yaml instead of raw bash script
- Tidy up the modules and replace deprecated call WithInsecure
- Add option to create config for rhc-worker-bash
- Fix build for go1.16
- Use separate environment for every executed command
- Update to make the yaml-file more generic
- Expected yaml structure should be list on top level
- Add setup for sos report
- Update the worker to make it more generic

* Thu Jul 06 2023 Eric Gustavsson <egustavs@redhat.com> 0.2-1
- Fix RPM specfile Source

* Wed Jun 14 2023 Rodolfo Olivieri <rolivier@redhat.com> 0.1-1
- Initial RPM release
