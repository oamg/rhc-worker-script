# Specfile based and updated from
# https://github.com/theforeman/foreman-packaging/blob/rpm/develop/packages/client/foreman_ygg_worker/foreman_ygg_worker.spec

%define debug_package %{nil}

%global repo_orgname oamg
%global repo_name rhc-worker-script
%global binary_name rhc-script-worker
%global rhc_libexecdir %{_libexecdir}/rhc
%{!?_root_sysconfdir:%global _root_sysconfdir %{_sysconfdir}}
%global rhc_worker_conf_dir %{_root_sysconfdir}/rhc/workers

%define gobuild(o:) env GO111MODULE=off go build -buildmode pie -compiler gc -tags="rpm_crashtraceback ${BUILDTAGS:-}" -ldflags "${LDFLAGS:-} -linkmode=external -B 0x$(head -c20 /dev/urandom|od -An -tx1|tr -d ' \\n') -extldflags '-Wl,-z,relro -Wl,-z,now -specs=/usr/lib/rpm/redhat/redhat-hardened-ld '" -a -v %{?**}

# EL7 doesn't define go_arches (it is available in go-srpm-macros which is EL8+)
%if !%{defined go_arches}
%define go_arches x86_64
%endif

%global use_go_toolset_1_19 0%{?rhel} == 7 && !%{defined centos}

Name:           %{repo_name}
Version:        0.7
Release:        1%{?dist}
Summary:        Worker executing scripts on hosts managed by Red Hat Insights

License:        GPLv3+
URL:            https://github.com/%{repo_orgname}/%{repo_name}
Source0:        %{url}/releases/download/v%{version}/%{name}-%{version}.tar.gz
ExclusiveArch:  %{go_arches}

%if %{use_go_toolset_1_19}
BuildRequires:  go-toolset-1.19-golang
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
mkdir -p _gopath/src
ln -fs $(pwd)/src _gopath/src/%{binary_name}-%{version}
ln -fs $(pwd)/vendor _gopath/src/%{binary_name}-%{version}/vendor
export GOPATH=$(pwd)/_gopath
pushd _gopath/src/%{binary_name}-%{version}
%if %{use_go_toolset_1_19}
scl enable go-toolset-1.19 -- %{gobuild}
%else
%{gobuild}
%endif
strip %{binary_name}-%{version}
popd


%install
# Create a temporary directory /var/lib/rhc-worker-script - used mainly for storing temporary files
install -d %{buildroot}%{_sharedstatedir}/%{binary_name}/

install -D -m 755 _gopath/src/%{binary_name}-%{version}/%{binary_name}-%{version} %{buildroot}%{rhc_libexecdir}/%{binary_name}
install -D -d -m 755 %{buildroot}%{rhc_worker_conf_dir}

cat <<EOF >%{buildroot}%{rhc_worker_conf_dir}/rhc-worker-script.yml
# recipient directive to register with dispatcher
directive: "%{name}"

# whether to verify incoming yaml files
verify_yaml: true

# temporary directory in which the temporary script will be placed and executed.
temporary_worker_directory: "/var/lib/rhc-worker-script"
EOF


%files
%{rhc_libexecdir}/%{binary_name}
%license LICENSE
%doc README.md
%config %{rhc_worker_conf_dir}/rhc-worker-script.yml

%changelog

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
