#!/usr/bin/env bash
set -e

cd "$(dirname "$BASH_SOURCE")"

# Downloads dependencies into vendor/ directory
mkdir -p vendor
cd vendor

git_clone_light() {
	pkg=$1
	bra=$2

	pkg_url=https://$pkg
	target_dir=src/$pkg

	echo "$pkg @ $bra: "

	if [ -d $target_dir ]; then
		echo "rm old, $pkg"
		rm -fr $target_dir
	fi

	echo "clone, $pkg"
	git clone --depth 1 --quiet --branch $bra  $pkg_url $target_dir
	echo "done"
}

git_clone() {
	pkg=$1
	rev=$2

	pkg_url=https://$pkg
	target_dir=src/$pkg

	echo "$pkg @ $rev: "

	if [ -d $target_dir ]; then
		echo "rm old, $pkg"
		rm -fr $target_dir
	fi

	echo "clone, $pkg"
	git clone --quiet --no-checkout $pkg_url $target_dir
	( cd $target_dir && git reset --quiet --hard $rev )
	echo "done"
}

go_get() {
	pkg=$1

	echo "go get $pkg"
	GOPATH=`pwd` go get $pkg
	echo "done"
}

go_get golang.org/x/exp/inotify
go_get github.com/mitchellh/goamz/s3
go_get github.com/mitchellh/goamz/aws

echo "don't forget to add vendor folder to your GOPATH (export GOPATH=\$GOPATH:\`pwd\`/vendor)"
