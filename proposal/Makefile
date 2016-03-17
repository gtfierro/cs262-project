.SUFFIXES : .pdf .tex .zip

LATEX=pdflatex -synctex=1 -interaction=nonstopmode --shell-escape

FLAGS=-shell-escape

TEXFILES = $(wildcard *.tex ../*.tex)

PDF = proposal.pdf

pdf: $(PDF)

%.pdf: %.tex $(TEXFILES)
	pdflatex $(FLAGS) $*
	bibtex $*
	pdflatex $(FLAGS) $*
	pdflatex $(FLAGS) $*

clean:
	rm -rf *~ *.log proposal.pdf *.aux *.out *.toc *.bbl *.blg
