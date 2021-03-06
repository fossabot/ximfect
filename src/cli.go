package main

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"strings"
	"ximfect/effect"
	"ximfect/environ"
	"ximfect/fxchain"
	"ximfect/libs"
	"ximfect/pack"
	"ximfect/tool"

	"github.com/ximfect/ximgy"
)

const (
	scriptTemplate string = "function effect(x, y, pixel) {\n	// write your code here\n	return {r: pixel.r, g: pixel.g, b: pixel.b, a: pixel.a};\n}\n"

	metaTemplate string = "name: Empty Effect\nversion: 1.0.0\nauthor: unknown <>\ndesc: ximfect generated empty effect\n"
)

var gTool *tool.Tool = tool.NewTool(
	"ximfect",
	tool.Version,
	"Learn more at https://ximfect.github.io")

func _version(t *tool.Tool, a tool.ArgumentList) error {
	t.PrintLn(t.GetVersion())
	return nil
}

func _apply(t *tool.Tool, a tool.ArgumentList) error {
	eff, hasEff := a.NamedArgs["effect"]
	file, hasFile := a.NamedArgs["file"]
	out, hasOut := a.NamedArgs["out"]

	if !hasEff {
		return errors.New(
			"missing effect argument, specify with --effect <id>")
	}
	if !hasFile {
		return errors.New(
			"missing input file, specify with --file <filename>")
	}
	if !hasOut {
		return errors.New(
			"missing output file, specify with --out <filename>")
	}

	effName := eff.Value
	inFileName := file.Value
	outFileName := out.Value

	t.VerboseLn("Loading effect:", effName)
	fx, err := effect.LoadFromAppdata(effName)
	if err != nil {
		return err
	}

	t.VerboseLn("Opening file:", inFileName)
	inFile, err := ximgy.Open(inFileName)
	if err != nil {
		return err
	}

	t.PrintLn("Applying effect:", effName)
	err = effect.Apply(fx, inFile, t, a)
	if err != nil {
		return err
	}

	t.VerboseLn("Saving output file:", outFileName)
	err = ximgy.Save(inFile, outFileName)
	if err != nil {
		return err
	}

	return nil
}

func _about(t *tool.Tool, a tool.ArgumentList) error {
	eff, hasEff := a.NamedArgs["effect"]
	lib, hasLib := a.NamedArgs["lib"]

	if !(hasEff || hasLib) {
		return errors.New(
			"what should be described? use --effect <id> or --lib <id>")
	}

	effName := strings.ToLower(eff.Value)
	libName := strings.ToLower(lib.Value)

	if hasEff {
		t.VerboseLn("Loading effect:", effName)
		fx, err := effect.LoadFromAppdata(effName)
		if err != nil {
			return err
		}

		meta := fx.Metadata

		t.PrintF("======== About %s ========\n", effName)
		t.PrintF("Name:           %s\n", meta.Name)
		t.PrintF("Version:        %s\n", meta.Version)
		t.PrintF("Author:         %s\n", meta.Author)
		t.PrintF("Description:    %s\n", meta.Desc)
		if len(meta.Preload) > 0 {
			t.PrintF("Preload:         %v\n", strings.Join(meta.Preload, ", "))
		}
	} else if hasLib {
		t.VerboseLn("Loading lib:", libName)
		library, err := libs.LoadFromAppdata(libName)
		if err != nil {
			return err
		}

		meta := library.Metadata

		fmt.Printf("======== About %s ========\n", libName)
		fmt.Printf("Name:           %s\n", meta.Name)
		fmt.Printf("Version:        %s\n", meta.Version)
		fmt.Printf("Author:         %s\n", meta.Author)
		fmt.Printf("Description:    %s\n", meta.Desc)
		fmt.Printf("Files:\n    [%s]\n", strings.Join(library.Files, "; "))
	}

	return nil
}

func _pack(t *tool.Tool, a tool.ArgumentList) error {
	eff, hasEff := a.NamedArgs["effect"]
	lib, hasLib := a.NamedArgs["lib"]

	if !hasEff && !hasLib {
		return errors.New(
			"missing input argument, specify with --effect <id> or --lib <id>")
	}

	if hasEff && hasLib {
		return errors.New(
			"too many input arguments, use ONLY --effect or ONLY --lib")
	}

	if hasEff {
		effName := strings.ToLower(eff.Value)
		outFileName := effName + ".fx.xpk"

		t.VerboseLn("Loading effect:", effName)
		_, err := effect.LoadFromAppdata(effName)
		if err != nil {
			return fmt.Errorf(
				"could not find effect: %s", effName)
		}

		t.PrintLn("Packaging...")
		path := environ.AppdataPath("effects", effName)
		raw, err := pack.GetPackedDirectory(path)
		if err != nil {
			return err
		}

		t.VerboseLn("Saving to file:", outFileName)
		file, err := os.Create(outFileName)
		if err != nil {
			return err
		}
		file.Write(raw)

		return nil
	} else {
		libName := strings.ToLower(lib.Value)
		outFileName := libName + ".lib.xpk"

		t.VerboseLn("Loading lib:", libName)
		_, err := libs.LoadFromAppdata(libName)
		if err != nil {
			return fmt.Errorf(
				"could not find effect: %s", libName)
		}

		t.VerboseLn("Packaging...")
		path := environ.AppdataPath("libs", libName)
		raw, err := pack.GetPackedDirectory(path)
		if err != nil {
			return err
		}

		t.VerboseLn("Saving to file:", outFileName)
		file, err := os.Create(outFileName)
		if err != nil {
			return err
		}
		file.Write(raw)

		t.VerboseLn("Finished!")
		return nil
	}
}

func _unpackEffect(t *tool.Tool, a tool.ArgumentList) error {
	file, hasFile := a.NamedArgs["file"]

	if !hasFile {
		return errors.New(
			"missing input file, specify with --file <filename>")
	}

	inFileName := file.Value

	t.VerboseLn("Reading file:", inFileName)
	raw, err := environ.LoadRawfile(inFileName)
	if err != nil {
		return err
	}
	//fmt.Println(raw)

	t.VerboseLn("Parsing package...")
	pkg, err := pack.GetPackage(raw)
	//fmt.Println(pkg)
	if err != nil {
		return err
	}

	t.PrintLn("Unpacking...")
	err = pack.UnpackTo(pkg, environ.AppdataPath("effects", pkg.Name))
	if err != nil {
		return err
	}

	return nil
}

func _unpackLib(t *tool.Tool, a tool.ArgumentList) error {
	file, hasFile := a.NamedArgs["file"]

	if !hasFile {
		return errors.New(
			"missing input file, specify with --file <filename>")
	}

	inFileName := file.Value

	t.VerboseLn("Reading file:", inFileName)
	raw, err := environ.LoadRawfile(inFileName)
	if err != nil {
		return err
	}

	t.VerboseLn("Parsing package...")
	pkg, err := pack.GetPackage(raw)
	if err != nil {
		return err
	}

	t.PrintLn("Unpacking...")
	err = pack.UnpackTo(pkg, environ.AppdataPath("libs", pkg.Name))
	if err != nil {
		return err
	}

	return nil
}

func _applyChain(t *tool.Tool, a tool.ArgumentList) error {
	file, hasFile := a.NamedArgs["file"]
	out, hasOut := a.NamedArgs["out"]
	inp, hasInp := a.NamedArgs["img"]

	if !hasFile {
		return errors.New(
			"missing input file, specify with --file <filename>")
	}
	if !hasOut {
		return errors.New(
			"missing output file, specify with --out <filename>")
	}
	if !hasInp {
		return errors.New(
			"missing input image, specify with --img <filename>")
	}

	inFileName := file.Value
	outFileName := out.Value
	inpFileName := inp.Value

	t.VerboseLn("Loading FX chain: ", inFileName)
	src, err := environ.LoadTextfile(inFileName)
	if err != nil {
		return err
	}

	t.VerboseLn("Loading image:", inpFileName)
	img, err := ximgy.Open(inpFileName)
	if err != nil {
		return err
	}

	t.PrintLn("Applying FX chain...")
	res, err := fxchain.Apply(src, img, t)
	if err != nil {
		return err
	}

	t.VerboseLn("Saving result:", outFileName)
	err = ximgy.Save(res, outFileName)
	if err != nil {
		return err
	}

	return nil
}

func _test(t *tool.Tool, a tool.ArgumentList) error {
	out, hasOut := a.NamedArgs["out"]

	if !hasOut {
		return errors.New(
			"missing output file, specify with --out <filename>")
	}

	outFileName := out.Value

	t.VerboseLn("Generating test image...")
	amt := 1024
	img := ximgy.MakeEmpty(image.Rect(0, 0, amt, amt))
	step := amt / 256
	img.Iterate(func(pixel ximgy.Pixel) (color.RGBA, error) {
		return color.RGBA{uint8(pixel.X / step), 0, uint8(pixel.Y / step), 255}, nil
	})

	t.VerboseLn("Saving output file:", outFileName)
	err := ximgy.Save(img, outFileName)
	if err != nil {
		return err
	}

	t.VerboseLn("Finished!")
	return nil
}

func _fxInit(t *tool.Tool, a tool.ArgumentList) error {
	eff, hasEff := a.NamedArgs["effect"]

	if !hasEff {
		return errors.New(
			"missing effect argument, specify with --effect <id>")
	}

	effName := strings.ToLower(eff.Value)

	t.PrintLn("Creating effect structure")
	err := os.Mkdir(environ.AppdataPath("effects", effName), os.ModePerm)
	if err != nil {
		return err
	}
	script, err := os.Create(environ.AppdataPath("effects", effName, "effect.js"))
	if err != nil {
		return err
	}
	meta, err := os.Create(environ.AppdataPath("effects", effName, "effect.yml"))
	if err != nil {
		return err
	}

	t.VerboseLn("Writing file templates...")
	_, err = script.WriteString(scriptTemplate)
	if err != nil {
		return err
	}
	_, err = meta.WriteString(metaTemplate)
	if err != nil {
		return err
	}

	t.PrintLn(" -- View your effect in:", environ.AppdataPath("effects", effName))
	return nil
}

func _dev(t *tool.Tool, a tool.ArgumentList) error {
	panic("hello")
	return nil
}

func main() {
	environ.EnsureAppdata()
	gTool.Init()
	gTool.AddAction("version", _version, "Shows the version")
	gTool.AddAction("apply", _apply, "Applies an effect")
	gTool.AddAction("about", _about, "Shows information about and effect")
	gTool.AddAction("pack", _pack, "Packs an effect into a zip archive")
	gTool.AddAction("unpack-effect", _unpackEffect, "Unpacks and installs an effect")
	gTool.AddAction("unpack-lib", _unpackLib, "Unpacks and installs a lib")
	gTool.AddAction("save-test", _test, "Generates and saves a test image")
	gTool.AddAction("make-empty", _fxInit, "Generates an effect template")
	gTool.AddAction("dev", _dev, "Action for internal testing")
	gTool.AddAction("apply-chain", _applyChain, "Applies an FX chain from a file.")

	var err error

	if len(os.Args) == 1 {
		err = gTool.RunAction([]string{"", "help"})
	} else if strings.HasSuffix(os.Args[1], ".fx.xpk") {
		err = gTool.RunAction([]string{"", "unpack-effect", "--file", os.Args[1]})
	} else if strings.HasSuffix(os.Args[1], ".lib.xpk") {
		err = gTool.RunAction([]string{"", "unpack-lib", "--file", os.Args[1]})
	} else if strings.HasSuffix(os.Args[1], ".xfc") {
		err = gTool.RunAction([]string{"", "apply-chain", "--file", os.Args[1]})
	} else {
		err = gTool.RunAction(os.Args)
	}

	if err != nil {
		gTool.ErrorExit("ERROR:", err)
	}
}
