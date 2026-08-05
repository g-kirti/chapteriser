package transcribe

type LanguageConfig struct {
	Keywords           map[string]bool
	StandaloneKeywords map[string]bool
	NumberWordTypes    map[string]int
	AndFollowsTypes    map[int]bool
}

var LanguageRegistry = map[string]LanguageConfig{
	"en": {
		Keywords: map[string]bool{
			"book": true, "chapter": true, "volume": true, "part": true,
		},
		StandaloneKeywords: map[string]bool{
			"introduction": true, "prologue": true, "epilogue": true, "foreword": true, "afterword": true,
		},
		NumberWordTypes: map[string]int{
			"zero": numUnit, "one": numUnit, "two": numUnit, "three": numUnit, "four": numUnit,
			"five": numUnit, "six": numUnit, "seven": numUnit, "eight": numUnit, "nine": numUnit,
			"ten": numUnit, "eleven": numUnit, "twelve": numUnit, "thirteen": numUnit, "fourteen": numUnit,
			"fifteen": numUnit, "sixteen": numUnit, "seventeen": numUnit, "eighteen": numUnit, "nineteen": numUnit,
			"twenty": numTens, "thirty": numTens, "forty": numTens, "fifty": numTens, "sixty": numTens,
			"seventy": numTens, "eighty": numTens, "ninety": numTens,
			"hundred": numScale, "thousand": numScale,
			"and": numAnd,
		},
		AndFollowsTypes: map[int]bool{
			numScale: true, // "one hundred and five"
		},
	},
	"es": {
		Keywords: map[string]bool{
			"libro": true, "capítulo": true, "capitulo": true, "volumen": true, "parte": true,
		},
		StandaloneKeywords: map[string]bool{
			"introducción": true, "introduccion": true, "prólogo": true, "prologo": true,
			"epílogo": true, "epilogo": true, "prefacio": true, "colofón": true, "colofon": true,
		},
		NumberWordTypes: map[string]int{
			"cero": numUnit, "uno": numUnit, "dos": numUnit, "tres": numUnit, "cuatro": numUnit,
			"cinco": numUnit, "seis": numUnit, "siete": numUnit, "ocho": numUnit, "nueve": numUnit,
			"diez": numUnit, "once": numUnit, "doce": numUnit, "trece": numUnit, "catorce": numUnit,
			"quince": numUnit, "dieciséis": numUnit, "dieciseis": numUnit, "diecisiete": numUnit,
			"dieciocho": numUnit, "diecinueve": numUnit,
			"veintiuno": numUnit, "veintidós": numUnit, "veintidos": numUnit,
			"veintitrés": numUnit, "veintitres": numUnit, "veinticuatro": numUnit,
			"veinticinco": numUnit, "veintiséis": numUnit, "veintiseis": numUnit,
			"veintisiete": numUnit, "veintiocho": numUnit, "veintinueve": numUnit,
			"veinte": numTens, "treinta": numTens, "cuarenta": numTens, "cincuenta": numTens,
			"sesenta": numTens, "setenta": numTens, "ochenta": numTens, "noventa": numTens,
			"cien": numScale, "ciento": numScale,
			"doscientos": numScale, "doscientas": numScale,
			"trescientos": numScale, "trescientas": numScale,
			"cuatrocientos": numScale, "cuatrocientas": numScale,
			"quinientos": numScale, "quinientas": numScale,
			"seiscientos": numScale, "seiscientas": numScale,
			"setecientos": numScale, "setecientas": numScale,
			"ochocientos": numScale, "ochocientas": numScale,
			"novecientos": numScale, "novecientas": numScale,
			"mil": numScale,
			"y":   numAnd,
		},
		AndFollowsTypes: map[int]bool{
			numTens: true, // "treinta y uno"
		},
	},
	"fr": {
		Keywords: map[string]bool{
			"livre": true, "chapitre": true, "volume": true, "partie": true,
		},
		StandaloneKeywords: map[string]bool{
			"introduction": true, "prologue": true, "épilogue": true, "epilogue": true,
			"avant-propos": true, "avantpropos": true, "postface": true,
		},
		NumberWordTypes: map[string]int{
			"zéro": numUnit, "zero": numUnit, "un": numUnit, "une": numUnit, "deux": numUnit,
			"trois": numUnit, "quatre": numUnit, "cinq": numUnit, "six": numUnit, "sept": numUnit,
			"huit": numUnit, "neuf": numUnit,
			"dix":  numTeenBase,
			"onze": numUnit, "douze": numUnit,
			"treize": numUnit, "quatorze": numUnit, "quinze": numUnit, "seize": numUnit,
			"dix-sept": numUnit, "dixsept": numUnit, "dix-huit": numUnit, "dixhuit": numUnit,
			"dix-neuf": numUnit, "dixneuf": numUnit,
			"vingt": numTens, "trente": numTens, "quarante": numTens, "cinquante": numTens,
			"soixante": numTens, "soixante-dix": numTens, "soixantedix": numTens,
			"quatre-vingts": numTens, "quatrevingts": numTens, "quatre-vingt": numTens,
			"quatre-vingt-dix": numTens, "quatrevingtdix": numTens,
			"cent": numScale, "mille": numScale,
			"et": numAnd,
		},
		AndFollowsTypes: map[int]bool{
			numTens: true, // "trente et un"
		},
	},
}
