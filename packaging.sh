#!/usr/bin/env sh

CGO_ENABLED=0 gox -os="darwin linux" -arch="amd64" -tags netgo -ldflags '-w -s'

dir=binaries
mkdir -p $dir;
rm -rf $dir/$f;

for f in breseq-rm-ctrl_*; do
    mkdir -p $dir/$f;
    mv $f $dir/$f;
    cd $dir/$f;
    mv $f $(echo $f | perl -pe 's/_[^\.]+//g');
    tar -zcf $f.tar.gz breseq-rm-ctrl*;
    mv *.tar.gz ../;
    cd ..;
    rm -rf $f;
    cd ..;
done;
