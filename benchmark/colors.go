// Copyright 2024 @ghkdqhrbals Authors
// This file is part of the gotybench library.
//
// The gotybench library is free software: you can redistribute it and/or modify
// it under the terms of the MIT License.
package benchmark

import "github.com/fatih/color"

var (
	green      = color.New(color.FgGreen)
	boldGreen  = color.New(color.FgGreen).Add(color.Bold)
	yellow     = color.New(color.FgHiYellow)
	boldYellow = color.New(color.FgHiYellow).Add(color.Bold)
	mg         = color.New(color.FgHiMagenta)
	boldMg     = color.New(color.FgHiMagenta).Add(color.Bold)
	red        = color.New(color.FgHiRed)
	boldRed    = color.New(color.FgHiRed).Add(color.Bold)
	boldCyan   = color.New(color.FgCyan).Add(color.Bold)
	cyan       = color.New(color.FgCyan)
)