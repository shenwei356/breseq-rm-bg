# breseq-rm-ctrl

This tools is used for finding mutations using [breseq](https://github.com/barricklab/breseq) for experiments with a control sample.

It makes breseq output file (`index.html`) more readable by removing mutations that are in control sample.

## Download

[Executable binary files](https://github.com/shenwei356/breseq-rm-ctrl/releases).

## Usage

    breseq-rm-ctrl ctrl/output/index.html sample1/output/index.html > sample1/output/index2.html

## Compatibility

breseq-rm-ctrl `v0.1.0` is tested for breseq version `0.33.2`, it should works for recent breseq versions.

## License

[MIT License](https://github.com/shenwei356/unikmer/blob/master/LICENSE)
