# Specfile based and updated from
# https://github.com/theforeman/foreman-packaging/blob/rpm/develop/packages/client/foreman_ygg_worker/foreman_ygg_worker.spec

%define debug_package %{nil}

%global repo_orgname oamg
%global repo_name rhc-bash-worker
%global yggdrasil_libexecdir %{_libexecdir}/yggdrasil
%{!?_root_sysconfdir:%global _root_sysconfdir %{_sysconfdir}}
%global yggdrasil_worker_conf_dir %{_root_sysconfdir}/yggdrasil/workers

%global goipath         github.com/%{repo_orgname}/%{repo_name}

%if 0%{?rhel} > 7 && ! 0%{?fedora}
%define gobuild(o:) \
        go build -buildmode pie -compiler gc -tags="rpm_crashtraceback libtrust_openssl ${BUILDTAGS:-}" -ldflags "${LDFLAGS:-} -linkmode=external -compressdwarf=false -B 0x$(head -c20 /dev/urandom|od -An -tx1|tr -d ' \\n') -extldflags '%__global_ldflags'" -a -v %{?**};
%else
%if ! 0%{?gobuild:1}
%define gobuild(o:) GO111MODULE=off go build -buildmode pie -compiler gc -tags="rpm_crashtraceback ${BUILDTAGS:-}" -ldflags "${LDFLAGS:-} -linkmode=external -B 0x$(head -c20 /dev/urandom|od -An -tx1|tr -d ' \\n') -extldflags '-Wl,-z,relro -Wl,-z,now -specs=/usr/lib/rpm/redhat/redhat-hardened-ld '" -a -v %{?**};
%endif
%endif

Name:           rhc-bash-worker
Version:        v0.1
Release:        1%{?dist}
Summary:        Experimental worker for Convert2RHEL.

License:        GPLv3+
URL:            https://github.com/%{repo_orgname}/%{repo_name}/
Source0:        https://github.com/%{repo_orgname}/%{repo_name}/releases/download/v%{version}/%{repo_name}-%{version}.tar.gz
BuildArch:      noarch

BuildRequires:  golang
Requires:       yggdrasil

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
# Create a temporary directory /var/lib/rhc-bash-worker - used mainly for storing temporary files
install -d %{buildroot}%{_sharedstatedir}/%{name}/

install -D -m 755 _gopath/src/%{name}-%{version}/%{name}-%{version} %{buildroot}%{yggdrasil_libexecdir}/%{name}
install -D -d -m 755 %{buildroot}%{yggdrasil_worker_conf_dir}

cat <<EOF >%{buildroot}%{yggdrasil_worker_conf_dir}/rhc-bash-worker.toml
exec = "%{yggdrasil_libexecdir}/%{name}"
protocol = "grpc"
env = []
EOF

%files
%{yggdrasil_libexecdir}/%{name}
%{yggdrasil_worker_conf_dir}/rhc-bash-worker.toml
%license LICENSE
%doc README.md

%changelog

* Mon Jun 12 2023 Rodolfo Olivieri <rolivier@redhat.com> - v0.1-1.20230612143302651634.package.as.copr.build.3.g153ca5f
- Update specfile (Rodolfo Olivieri)
- Add LICENSE (Rodolfo Olivieri)
- Build with packit (Rodolfo Olivieri)

* Mon Jun 12 2023 Rodolfo Olivieri <rolivier@redhat.com> - v0.1-1.20230612143251359449.package.as.copr.build.3.g153ca5f
- Update specfile (Rodolfo Olivieri)
- Add LICENSE (Rodolfo Olivieri)
- Build with packit (Rodolfo Olivieri)

* Mon Jun 12 2023 Rodolfo Olivieri <rolivier@redhat.com> - v0.1-1.20230612143222265715.package.as.copr.build.3.g153ca5f
- Update specfile (Rodolfo Olivieri)
- Add LICENSE (Rodolfo Olivieri)
- Build with packit (Rodolfo Olivieri)

* Mon Jun 12 2023 Rodolfo Olivieri <rolivier@redhat.com> - v0.1-1.20230612143057210246.package.as.copr.build.3.g153ca5f
- Update specfile (Rodolfo Olivieri)
- Add LICENSE (Rodolfo Olivieri)
- Build with packit (Rodolfo Olivieri)

* Mon Jun 12 2023 Rodolfo Olivieri <rolivier@redhat.com> 0.1-1
- Initial RPM release
