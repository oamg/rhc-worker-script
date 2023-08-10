# Specfile based and updated from
# https://github.com/theforeman/foreman-packaging/blob/rpm/develop/packages/client/foreman_ygg_worker/foreman_ygg_worker.spec

%define debug_package %{nil}

%global repo_orgname oamg
%global repo_name rhc-worker-script
%global rhc_libexecdir %{_libexecdir}/rhc
%{!?_root_sysconfdir:%global _root_sysconfdir %{_sysconfdir}}
%global rhc_worker_conf_dir %{_root_sysconfdir}/rhc/workers

%define gobuild(o:) env GO111MODULE=off go build -buildmode pie -compiler gc -tags="rpm_crashtraceback ${BUILDTAGS:-}" -ldflags "${LDFLAGS:-} -linkmode=external -B 0x$(head -c20 /dev/urandom|od -An -tx1|tr -d ' \\n') -extldflags '-Wl,-z,relro -Wl,-z,now -specs=/usr/lib/rpm/redhat/redhat-hardened-ld '" -a -v %{?**}

# EL7 doesn't define go_arches (it is available in go-srpm-macros which is EL8+)
%if !%{defined go_arches}
%define go_arches x86_64 s390x ppc64le
%endif

%global use_go_toolset_1_16 0%{?rhel} == 7 && !%{defined centos}

Name:           %{repo_name}
Version:        0.3
Release:        1%{?dist}
Summary:        Worker executing scripts on hosts managed by Red Hat Insights

License:        GPLv3+
URL:            https://github.com/%{repo_orgname}/%{repo_name}
Source0:        %{url}/archive/v%{version}/%{name}-%{version}.tar.gz
ExclusiveArch:  %{go_arches}

%if %{use_go_toolset_1_16}
BuildRequires:  go-toolset-1.16-golang
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
ln -fs $(pwd)/src _gopath/src/%{repo_name}-%{version}
ln -fs $(pwd)/vendor _gopath/src/%{repo_name}-%{version}/vendor
export GOPATH=$(pwd)/_gopath
pushd _gopath/src/%{repo_name}-%{version}
%if %{use_go_toolset_1_16}
scl enable go-toolset-1.16 -- %{gobuild}
%else
%{gobuild}
%endif
strip %{repo_name}-%{version}
popd


%install
# Create a temporary directory /var/lib/rhc-worker-script - used mainly for storing temporary files
install -d %{buildroot}%{_sharedstatedir}/%{repo_name}/

install -D -m 755 _gopath/src/%{repo_name}-%{version}/%{repo_name}-%{version} %{buildroot}%{rhc_libexecdir}/%{repo_name}
install -D -d -m 755 %{buildroot}%{rhc_worker_conf_dir}

%files
%{rhc_libexecdir}/%{repo_name}
%license LICENSE
%doc README.md

%changelog

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
