# breseq-rm-bg

This tools is used for finding mutations using [breseq](https://github.com/barricklab/breseq) for experiments with control/background samples.

If you use breseq to find mutation in bacteria, but the reference strain is not well assembled or annotated,
you should re-seq your reference strain as a control/background, and use a close strain as reference for analysis.

`breseq-rm-bg` makes breseq output file (`output/index.html`) more readable by removing mutations that are in control/background sample.

## Download

[Executable binary files](https://github.com/shenwei356/breseq-rm-bg/releases).

## Usage

    breseq-rm-bg --bg-inter --bg-files bg1/output/index.html ---bg-files bg2/output/index.html \
      sample1/output/index.html > sample1/output/index.filtered.html

## Compatibility

breseq-rm-bg `v0.2.0` is tested for breseq version `0.33.2`, it should works for recent breseq versions.

## License

[MIT License](https://github.com/shenwei356/breseq-rm-bg/blob/master/LICENSE)
