.SUFFIXES : .pdf .tex .zip

LATEX=pdflatex -synctex=1 -interaction=nonstopmode --shell-escape

FLAGS=-shell-escape

TEXFILES = $(wildcard *.tex ../*.tex)

PDF = paper.pdf

pdf: $(PDF)

%.pdf: %.tex $(TEXFILES)
	pdflatex $(FLAGS) $*
	bibtex $*
	pdflatex $(FLAGS) $*
	pdflatex $(FLAGS) $*

opt: $(PDF)
	gs -q -dBATCH -dNOPAUSE -dSAFER -sDEVICE=pdfwrite \
	     -dPDFSETTINGS=/printer -dUseCIEColor -sOutputFile=out.pdf paper.pdf
	mv out.pdf paper.pdf

clean:
	rm -rf *~ *.log paper.pdf *.aux *.out *.toc *.bbl *.blg
