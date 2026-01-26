package basic

import (
	"testing"
)

func TestLetStatement(t *testing.T) {
	state := NewBasicState()
	state = StoreLine(state, 10, "LET A = 5")

	code := CompileProgram(state)
	if len(code) == 0 {
		t.Error("Expected code to be generated for LET statement")
	}

	// Verify code contains LDA #$05 (0xA9 0x05) and STA $10 (0x85 0x10)
	// LDA immediate = 0xA9, STA zero page = 0x85
	foundLDA := false
	foundSTA := false
	for i := 0; i < len(code)-1; i++ {
		if code[i] == 0xA9 && code[i+1] == 0x05 {
			foundLDA = true
		}
		if code[i] == 0x85 && code[i+1] == 0x10 {
			foundSTA = true
		}
	}

	if !foundLDA {
		t.Error("Expected LDA #$05 in generated code")
	}
	if !foundSTA {
		t.Error("Expected STA $10 in generated code")
	}
}

func TestGotoStatement(t *testing.T) {
	state := NewBasicState()
	state = StoreLine(state, 10, "GOTO 30")
	state = StoreLine(state, 20, "LET A = 1")
	state = StoreLine(state, 30, "LET B = 2")

	code := CompileProgram(state)
	if len(code) == 0 {
		t.Error("Expected code to be generated for GOTO program")
	}

	// Verify JMP instruction exists (0x4C)
	foundJMP := false
	for i := 0; i < len(code); i++ {
		if code[i] == 0x4C {
			foundJMP = true
			break
		}
	}

	if !foundJMP {
		t.Error("Expected JMP instruction in generated code")
	}
}

func TestForNextLoop(t *testing.T) {
	state := NewBasicState()
	state = StoreLine(state, 10, "FOR I = 1 TO 5")
	state = StoreLine(state, 20, "LET A = I")
	state = StoreLine(state, 30, "NEXT I")

	code := CompileProgram(state)
	if len(code) == 0 {
		t.Error("Expected code to be generated for FOR/NEXT loop")
	}

	// Verify INC instruction exists (0xE6 for zero page)
	foundINC := false
	for i := 0; i < len(code); i++ {
		if code[i] == 0xE6 {
			foundINC = true
			break
		}
	}

	if !foundINC {
		t.Error("Expected INC instruction in generated code for NEXT")
	}
}

func TestIfThenStatement(t *testing.T) {
	state := NewBasicState()
	state = StoreLine(state, 10, "LET A = 10")
	state = StoreLine(state, 20, "IF A > 5 THEN LET B = 1")

	code := CompileProgram(state)
	if len(code) == 0 {
		t.Error("Expected code to be generated for IF/THEN statement")
	}

	// Verify CMP instruction exists (0xC5 for zero page compare or 0xC9 for immediate)
	foundCMP := false
	for i := 0; i < len(code); i++ {
		if code[i] == 0xC5 || code[i] == 0xC9 {
			foundCMP = true
			break
		}
	}

	if !foundCMP {
		t.Error("Expected CMP instruction in generated code for IF")
	}
}

func TestGosubReturn(t *testing.T) {
	state := NewBasicState()
	state = StoreLine(state, 10, "GOSUB 100")
	state = StoreLine(state, 20, "END")
	state = StoreLine(state, 100, "LET A = 1")
	state = StoreLine(state, 110, "RETURN")

	code := CompileProgram(state)
	if len(code) == 0 {
		t.Error("Expected code to be generated for GOSUB/RETURN")
	}

	// Verify JSR (0x20) and RTS (0x60) instructions exist
	foundJSR := false
	foundRTS := false
	for i := 0; i < len(code); i++ {
		if code[i] == 0x20 {
			foundJSR = true
		}
		if code[i] == 0x60 {
			foundRTS = true
		}
	}

	if !foundJSR {
		t.Error("Expected JSR instruction in generated code")
	}
	if !foundRTS {
		t.Error("Expected RTS instruction in generated code")
	}
}

func TestVariableAddress(t *testing.T) {
	// Test variable A = $10
	if GetVariableAddress("A") != 0x10 {
		t.Errorf("Expected A to be at $10, got $%02X", GetVariableAddress("A"))
	}
	// Test variable Z = $29
	if GetVariableAddress("Z") != 0x29 {
		t.Errorf("Expected Z to be at $29, got $%02X", GetVariableAddress("Z"))
	}
	// Test lowercase
	if GetVariableAddress("a") != 0x10 {
		t.Errorf("Expected lowercase a to be at $10, got $%02X", GetVariableAddress("a"))
	}
}

func TestExpressionParsing(t *testing.T) {
	// Test simple number
	expr := ParseExpression("5")
	if expr.Type != ExprNumber || expr.Value != 5 {
		t.Error("Failed to parse number expression")
	}

	// Test variable
	expr = ParseExpression("A")
	if expr.Type != ExprVariable || expr.VarName != "A" {
		t.Error("Failed to parse variable expression")
	}

	// Test addition
	expr = ParseExpression("3 + 2")
	if expr.Type != ExprBinaryOp || expr.Op != "+" {
		t.Error("Failed to parse addition expression")
	}
}
