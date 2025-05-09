#!/bin/bash
set -e
# Copyright 2017 The OpenEBS Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

SRC_REPO="$TRAVIS_BUILD_DIR"
DST_REPO="$GOPATH/src/github.com/aamir-tiwari-sumo/maya"

function checkGitDiff() {
	if [[ `git diff --shortstat | wc -l` != 0 ]]; then echo "Some files got changed after $1";printf "\n";git diff --stat;printf "\n"; exit 1; fi
}

if [ "$SRC_REPO" != "$DST_REPO" ];
then
	echo "Copying from $SRC_REPO to $DST_REPO"
	# Get the git commit
	echo "But first, get the git revision from $SRC_REPO"
	GIT_COMMIT="$(git rev-parse HEAD)"
	echo $GIT_COMMIT >> $SRC_REPO/GITCOMMIT

	mkdir -p $DST_REPO
	rsync -az ${TRAVIS_BUILD_DIR}/ ${DST_REPO}/
	export TRAVIS_BUILD_DIR=$DST_REPO
	cd $DST_REPO
fi

#Run common checks
make check-license
rc=$?; if [[ $rc != 0 ]]; then exit $rc; fi

echo "Running : make format"
make format
rc=$?; if [[ $rc != 0 ]]; then echo "make format failed"; exit $rc; fi
checkGitDiff "make format"
printf "\n"

echo "Running : verify module dependencies"
GO111MODULE=on make verify-deps

if [ "$TRAVIS_CPU_ARCH" == "amd64" ]; then
  # kubegen and unit tests are executed only for amd64
  echo "Running : make kubegen"
  make kubegen
  rc=$?; if [[ $rc != 0 ]]; then echo "make kubegen failed"; exit $rc; fi
  checkGitDiff "make kubegen"
  printf "\n"

  ./buildscripts/test-cov.sh
  rc=$?; if [[ $rc != 0 ]]; then exit $rc; fi

  make all
  rc=$?; if [[ $rc != 0 ]]; then exit $rc; fi

elif [ "$TRAVIS_CPU_ARCH" == "arm64" ]; then
  make all.arm64
  rc=$?; if [[ $rc != 0 ]]; then exit $rc; fi

elif [ "$TRAVIS_CPU_ARCH" == "ppc64le" ]; then
  make all.ppc64le
  rc=$?; if [[ $rc != 0 ]]; then exit $rc; fi
fi

if [ $SRC_REPO != $DST_REPO ] && [ -f "coverage.txt" ];
then
	echo "Copying coverage.txt to $SRC_REPO"
	cp coverage.txt $SRC_REPO/
	cd $SRC_REPO
fi
