%bcond_without check

# TODO(r0x0d): Update this reference when we move to @oamg
%global goipath         github.com/r0x0d/rhc-bash-worker

%gometa -f

%global godocs          CONTRIBUTING.md README.md

%global common_description %{expand:
Experimental worker for Convert2RHEL.}

Name:           rhc-bash-worker
Version:        v0.1
Release:        1%{?dist}
Summary:        Experimental worker for Convert2RHEL.

License:        # FIXME
URL:            %{gourl}
Source:         %(gosource)
BuildArch:      noarch

%description %{common_description}

%gopkg

%prep
%goprep

%generate_buildrequires
%go_generate_buildrequires

%build
for cmd in src; do
  %gobuild -o %{gobuilddir}/bin/$(basename $cmd) %{goipath}/$cmd
done

%install
%gopkginstall

# Create a temporary directory /var/lib/rhc-bash-worker - used mainly for storing temporary files
install -d %{buildroot}%{_sharedstatedir}/%{name}/
install -m 0755 -vd                     %{buildroot}%{_bindir}
# install -m 0755 -vp %{gobuilddir}/bin/* %{buildroot}%{_bindir}/
install -m 0755 -vp %{buildroot}%{_libexecdir}/yggdrasil/%{_bindir}/

%if %{with check}
%check
%gocheck
%endif

%files
%doc CONTRIBUTING.md README.md
%{_bindir}/*

%gopkgfiles

%changelog
%autochangelog
