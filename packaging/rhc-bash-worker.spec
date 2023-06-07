Name:           rhc-bash-worker
Version:        0.1
Release:        1%{?dist}
Summary:        RHC bash worker

License:
URL:
Source0:        %{name}-%{version}.tar.gz

BuildRequires:  golang
Requires:       system-rpm-macros

Provides:       %{name} = %{version}

%description
test

%prep
%autosetup


%build
CGO_ENABLED=0 go build -v -o %{name}

%configure
%make_build


%install
%make_install


%files
%license add-license-file-here
%doc add-docs-here



%changelog
* Wed May 24 2023 r0x0d
-
