# Specfile based and updated from
# https://github.com/theforeman/foreman-packaging/blob/rpm/develop/packages/client/foreman_ygg_worker/foreman_ygg_worker.spec

%define debug_package %{nil}

%global repo_orgname oamg
%global repo_name rhc-worker-bash
%global rhc_libexecdir %{_libexecdir}/rhc
%{!?_root_sysconfdir:%global _root_sysconfdir %{_sysconfdir}}
%global rhc_worker_conf_dir %{_root_sysconfdir}/rhc/workers

%define gobuild(o:) GO111MODULE=off go build -buildmode pie -compiler gc -tags="rpm_crashtraceback ${BUILDTAGS:-}" -ldflags "${LDFLAGS:-} -linkmode=external -B 0x$(head -c20 /dev/urandom|od -An -tx1|tr -d ' \\n') -extldflags '-Wl,-z,relro -Wl,-z,now -specs=/usr/lib/rpm/redhat/redhat-hardened-ld '" -a -v %{?**};

# EL7 doesn't define go_arche (it is available in go-srpm-macros which is EL8+)s
%if ! 0%{?go_arches:1}
%define go_arches %{ix86} x86_64 %{arm} aarch64 ppc64le
%endif

Name:           rhc-bash-worker
Version:        0.1
Release:        1%{?dist}
Summary:        Experimental worker for Convert2RHEL.

License:        GPLv3+
URL:            https://github.com/%{repo_orgname}/%{repo_name}/
Source0:        https://github.com/%{repo_orgname}/%{repo_name}/releases/download/v%{version}/%{repo_name}-%{version}.tar.gz
ExclusiveArch:  %{go_arches}

BuildRequires:  golang
Requires:       rhc

%description
Experimental worker for Convert2RHEL.

%prep
%setup -q

%build
mkdir -p _gopath/src
ln -fs $(pwd)/src _gopath/src/%{name}-%{version}
ln -fs $(pwd)/vendor _gopath/src/%{name}-%{version}/vendor
export GOPATH=$(pwd)/_gopath
pushd _gopath/src/%{name}-%{version}
%{gobuild}
strip %{name}-%{version}
popd


%install
# Create a temporary directory /var/lib/rhc-worker-bash - used mainly for storing temporary files
install -d %{buildroot}%{_sharedstatedir}/%{name}/

install -D -m 755 _gopath/src/%{name}-%{version}/%{name}-%{version} %{buildroot}%{rhc_libexecdir}/%{name}
install -D -d -m 755 %{buildroot}%{rhc_worker_conf_dir}

%files
%{rhc_libexecdir}/%{name}
%license LICENSE
%doc README.md

%changelog

* Wed Jun 14 2023 Rodolfo Olivieri <rolivier@redhat.com> 0.1-1
- Initial RPM release
