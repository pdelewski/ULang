// ASFormatter.cpp
// Copyright (c) 2025 The Artistic Style Authors.
// This code is licensed under the MIT License.
// License.md describes the conditions under which this software may be distributed.

//-----------------------------------------------------------------------------
// headers
//-----------------------------------------------------------------------------

#include "astyle.h"

#include <algorithm>
#include <fstream>
#include <set>
#include <string>


//-----------------------------------------------------------------------------
// astyle namespace
//-----------------------------------------------------------------------------

namespace astyle {

//
//-----------------------------------------------------------------------------
// ASFormatter class
//-----------------------------------------------------------------------------

/**
 * Constructor of ASFormatter
 */
ASFormatter::ASFormatter()
{
	sourceIterator = nullptr;
	enhancer = new ASEnhancer;
	preBraceHeaderStack = nullptr;
	braceTypeStack = nullptr;
	parenStack = nullptr;
	structStack = nullptr;
	questionMarkStack = nullptr;
	lineCommentNoIndent = false;
	formattingStyle = STYLE_NONE;
	braceFormatMode = NONE_MODE;
	pointerAlignment = PTR_ALIGN_NONE;
	referenceAlignment = REF_SAME_AS_PTR;
	objCColonPadMode = COLON_PAD_NO_CHANGE;
	lineEnd = LINEEND_DEFAULT;
	squeezeEmptyLineNum = std::string::npos;
	maxCodeLength = std::string::npos;
	isInStruct = false;
	shouldPadCommas = false;
	shouldPadOperators = false;
	negationPadMode = NEGATION_PAD_NO_CHANGE;
	includeDirectivePaddingMode = INCLUDE_PAD_NO_CHANGE;
	shouldPadParensOutside = false;
	shouldPadFirstParen = false;
	shouldPadEmptyParens = false;
	shouldPadParensInside = false;
	shouldPadHeader = false;
	shouldStripCommentPrefix = false;
	shouldUnPadParens = false;
	attachClosingBraceMode = false;
	shouldBreakOneLineBlocks = true;
	shouldBreakOneLineHeaders = false;
	shouldBreakOneLineStatements = true;
	shouldConvertTabs = false;
	shouldIndentCol1Comments = false;
	shouldIndentPreprocBlock = false;
	shouldCloseTemplates = false;
	shouldAttachExternC = false;
	shouldAttachNamespace = false;
	shouldAttachClass = false;
	shouldAttachClosingWhile = false;
	shouldAttachInline = false;
	shouldBreakBlocks = false;
	shouldBreakClosingHeaderBlocks = false;
	shouldBreakClosingHeaderBraces = false;
	shouldDeleteEmptyLines = false;
	shouldBreakReturnType = false;
	shouldBreakReturnTypeDecl = false;
	shouldAttachReturnType = false;
	shouldAttachReturnTypeDecl = false;
	shouldBreakElseIfs = false;
	shouldBreakLineAfterLogical = false;
	shouldAddBraces = false;
	shouldAddOneLineBraces = false;
	shouldRemoveBraces = false;
	shouldPadMethodColon = false;
	shouldPadMethodPrefix = false;
	shouldUnPadMethodPrefix = false;
	shouldPadReturnType = false;
	shouldUnPadReturnType = false;
	shouldPadParamType = false;
	shouldUnPadParamType = false;
	shouldPadBracketsOutside = false;
	shouldPadBracketsInside = false;
	shouldUnPadBrackets = false;
	isInMultlineStatement = false;
	isInExplicitBlock = 0;

	// initialize ASFormatter member std::vectors
	formatterFileType = INVALID_TYPE;		// reset to an invalid type
	headers = new std::vector<const std::string*>;
	nonParenHeaders = new std::vector<const std::string*>;
	preDefinitionHeaders = new std::vector<const std::string*>;
	preCommandHeaders = new std::vector<const std::string*>;
	operators = new std::vector<const std::string*>;
	assignmentOperators = new std::vector<const std::string*>;
	castOperators = new std::vector<const std::string*>;

	// initialize ASEnhancer member std::vectors
	indentableMacros = new std::vector<const std::pair<const std::string, const std::string>* >;
}

/**
 * Destructor of ASFormatter
 */
ASFormatter::~ASFormatter()
{
	// delete ASFormatter stack std::vectors
	deleteContainer(preBraceHeaderStack);
	deleteContainer(braceTypeStack);
	deleteContainer(parenStack);
	deleteContainer(structStack);
	deleteContainer(questionMarkStack);

	// delete ASFormatter member std::vectors
	formatterFileType = INVALID_TYPE;		// reset to an invalid type
	delete headers;
	delete nonParenHeaders;
	delete preDefinitionHeaders;
	delete preCommandHeaders;
	delete operators;
	delete assignmentOperators;
	delete castOperators;

	// delete ASEnhancer member std::vectors
	delete indentableMacros;

	// must be done when the ASFormatter object is deleted (not ASBeautifier)
	// delete ASBeautifier member std::vectors
	ASBeautifier::deleteBeautifierVectors();

	delete enhancer;
}

/**
 * initialize the ASFormatter.
 *
 * init() should be called every time a ASFormatter object is to start
 * formatting a NEW source file.
 * init() receives a pointer to a ASSourceIterator object that will be
 * used to iterate through the source code.
 *
 * @param si        a pointer to the ASSourceIterator or ASStreamIterator object.
 */
void ASFormatter::init(ASSourceIterator* si)
{
	buildLanguageVectors();
	fixOptionVariableConflicts();
	ASBeautifier::init(si);
	sourceIterator = si;

	enhancer->init(getFileType(),
	               getIndentLength(),
	               getTabLength(),
	               getIndentString() == "\t",
	               getForceTabIndentation(),
	               getNamespaceIndent(),
	               getCaseIndent(),
	               shouldIndentPreprocBlock,
	               getPreprocDefineIndent(),
	               getEmptyLineFill(),
	               indentableMacros);

	initContainer(preBraceHeaderStack, new std::vector<const std::string*>);
	initContainer(parenStack, new std::vector<int>);
	initContainer(structStack, new std::vector<bool>);
	initContainer(questionMarkStack, new std::vector<bool>);
	parenStack->emplace_back(0);               // parenStack must contain this default entry
	initContainer(braceTypeStack, new std::vector<BraceType>);
	braceTypeStack->emplace_back(NULL_TYPE);   // braceTypeStack must contain this default entry
	clearFormattedLineSplitPoints();

	currentHeader = nullptr;
	previousHeader = nullptr;
	currentLine = "";
	readyFormattedLine = "";
	formattedLine = "";
	verbatimDelimiter = "";
	currentChar = ' ';
	previousChar = ' ';
	previousCommandChar = ' ';
	previousNonWSChar = ',';	// not a potential name or operator
	quoteChar = '"';
	preprocBlockEnd = 0;
	charNum = 0;
	checksumIn = 0;
	checksumOut = 0;
	currentLineFirstBraceNum = std::string::npos;
	formattedLineCommentNum = 0;
	leadingSpaces = 0;
	previousReadyFormattedLineLength = std::string::npos;
	preprocBraceTypeStackSize = 0;
	spacePadNum = 0;
	methodAttachCharNum = std::string::npos;
	methodAttachLineNum = 0;
	methodBreakCharNum = std::string::npos;
	methodBreakLineNum = 0;
	nextLineSpacePadNum = 0;
	objCColonAlign = 0;
	templateDepth = 0;
	squareBracketCount = 0;
	parenthesesCount = 0;
	bracesNestingLevel = 0;
	bracesNestingLevelOfStruct = 0;
	squeezeEmptyLineCount = 0;

	runInIndentChars = 0;
	tabIncrementIn = 0;
	previousBraceType = NULL_TYPE;

	isVirgin = true;
	isInVirginLine = true;
	isInLineComment = false;
	isInComment = false;
	isInCommentStartLine = false;
	noTrimCommentContinuation = false;
	isInPreprocessor = false;
	isInPreprocessorDefineDef = false;
	isInPreprocessorBeautify = false;
	doesLineStartComment = false;
	lineEndsInCommentOnly = false;
	lineIsCommentOnly = false;
	lineIsLineCommentOnly = false;
	lineIsEmpty = false;
	isImmediatelyPostCommentOnly = false;
	isImmediatelyPostEmptyLine = false;
	isInClassInitializer = false;
	isInQuote = false;
	isInVerbatimQuote = false;
	checkInterpolation = false;
	haveLineContinuationChar = false;
	isInQuoteContinuation = false;
	isHeaderInMultiStatementLine = false;
	isSpecialChar = false;
	isNonParenHeader = false;
	foundNamespaceHeader = false;
	foundClassHeader = false;
	foundStructHeader = false;
	foundInterfaceHeader = false;
	foundPreDefinitionHeader = false;
	foundPreCommandHeader = false;
	foundPreCommandMacro = false;
	foundTrailingReturnType = false;
	foundCastOperator = false;
	foundQuestionMark = false;
	isInLineBreak = false;
	endOfAsmReached = false;
	endOfCodeReached = false;
	isFormattingModeOff = false;
	isInEnum = false;
	isInContinuedPreProc = false;
	isInStruct = false;
	isInExecSQL = false;
	isInAsm = false;
	isInAsmOneLine = false;
	isInAsmBlock = false;
	isLineReady = false;
	elseHeaderFollowsComments = false;
	caseHeaderFollowsComments = false;
	isPreviousBraceBlockRelated = false;
	isInPotentialCalculation = false;
	needHeaderOpeningBrace = false;
	shouldBreakLineAtNextChar = false;
	shouldKeepLineUnbroken = false;
	shouldReparseCurrentChar = false;
	passedSemicolon = false;
	passedColon = false;
	isImmediatelyPostNonInStmt = false;
	isCharImmediatelyPostNonInStmt = false;
	isInTemplate = false;
	isImmediatelyPostComment = false;
	isImmediatelyPostLineComment = false;
	isImmediatelyPostEmptyBlock = false;
	isImmediatelyPostObjCMethodPrefix = false;
	isImmediatelyPostPreprocessor = false;
	isImmediatelyPostReturn = false;
	isImmediatelyPostThrow = false;
	isImmediatelyPostNewDelete = false;
	isImmediatelyPostOperator = false;
	isImmediatelyPostTemplate = false;
	isImmediatelyPostPointerOrReference = false;
	isCharImmediatelyPostReturn = false;
	isCharImmediatelyPostThrow = false;
	isCharImmediatelyPostNewDelete = false;
	isCharImmediatelyPostOperator = false;
	isCharImmediatelyPostComment = false;
	isPreviousCharPostComment = false;
	isCharImmediatelyPostLineComment = false;
	isCharImmediatelyPostOpenBlock = false;
	isCharImmediatelyPostCloseBlock = false;
	isCharImmediatelyPostTemplate = false;
	isCharImmediatelyPostPointerOrReference = false;
	isInObjCInterface = false;
	isInObjCMethodDefinition = false;
	isInObjCReturnType = false;
	isInObjCParam = false;
	isInObjCSelector = false;
	breakCurrentOneLineBlock = false;
	shouldRemoveNextClosingBrace = false;
	isInBraceRunIn = false;
	returnTypeChecked = false;
	currentLineBeginsWithBrace = false;
	isPrependPostBlockEmptyLineRequested = false;
	isAppendPostBlockEmptyLineRequested = false;
	isIndentablePreprocessor = false;
	isIndentablePreprocessorBlck = false;
	prependEmptyLine = false;
	appendOpeningBrace = false;
	foundClosingHeader = false;
	isImmediatelyPostHeader = false;
	isInHeader = false;
	isInCase = false;
	isInAllocator = false;
	isInMultlineStatement = false;
	isInExplicitBlock = 0;

	isFirstPreprocConditional = false;
	processedFirstConditional = false;
	isJavaStaticConstructor = false;
}

/**
 * build std::vectors for each programming language
 * depending on the file extension.
 */
void ASFormatter::buildLanguageVectors()
{
	if (getFileType() == formatterFileType)  // don't build unless necessary
		return;

	formatterFileType = getFileType();

	headers->clear();
	nonParenHeaders->clear();
	preDefinitionHeaders->clear();
	preCommandHeaders->clear();
	operators->clear();
	assignmentOperators->clear();
	castOperators->clear();
	indentableMacros->clear();	// ASEnhancer

	ASResource::buildHeaders(headers, formatterFileType);
	ASResource::buildNonParenHeaders(nonParenHeaders, formatterFileType);
	ASResource::buildPreDefinitionHeaders(preDefinitionHeaders, formatterFileType);
	ASResource::buildPreCommandHeaders(preCommandHeaders, formatterFileType);
	ASResource::buildOperators(operators, formatterFileType);
	ASResource::buildAssignmentOperators(assignmentOperators);
	ASResource::buildCastOperators(castOperators);
	ASResource::buildIndentableMacros(indentableMacros);	//ASEnhancer
}

/**
 * set the variables for each predefined style.
 * this will override any previous settings.
 */
void ASFormatter::fixOptionVariableConflicts()
{
	if (formattingStyle == STYLE_ALLMAN)
	{
		setBraceFormatMode(BREAK_MODE);
	}
	else if (formattingStyle == STYLE_JAVA)
	{
		setBraceFormatMode(ATTACH_MODE);
	}
	else if (formattingStyle == STYLE_KR)
	{
		setBraceFormatMode(LINUX_MODE);
	}
	else if (formattingStyle == STYLE_STROUSTRUP)
	{
		setBraceFormatMode(LINUX_MODE);
		setBreakClosingHeaderBracesMode(true);
	}
	else if (formattingStyle == STYLE_WHITESMITH)
	{
		setBraceFormatMode(BREAK_MODE);
		setBraceIndent(true);
		setClassIndent(true);			// avoid hanging indent with access modifiers
		setSwitchIndent(true);			// avoid hanging indent with case statements
	}
	else if (formattingStyle == STYLE_VTK)
	{
		// the unindented class brace does NOT cause a hanging indent like Whitesmith
		setBraceFormatMode(BREAK_MODE);
		setBraceIndentVtk(true);		// sets both braceIndent and braceIndentVtk
		setSwitchIndent(true);			// avoid hanging indent with case statements
	}
	else if (formattingStyle == STYLE_RATLIFF)
	{
		// attached braces can have hanging indents with the closing brace
		setBraceFormatMode(ATTACH_MODE);
		setBraceIndent(true);
		setClassIndent(true);			// avoid hanging indent with access modifiers
		setSwitchIndent(true);			// avoid hanging indent with case statements
	}
	else if (formattingStyle == STYLE_GNU)
	{
		setBraceFormatMode(BREAK_MODE);
		setBlockIndent(true);
	}
	else if (formattingStyle == STYLE_LINUX)
	{
		setBraceFormatMode(LINUX_MODE);
		// always for Linux style
		setMinConditionalIndentOption(MINCOND_ONEHALF);
	}
	else if (formattingStyle == STYLE_HORSTMANN)
	{
		setBraceFormatMode(RUN_IN_MODE);
		setSwitchIndent(true);
	}
	else if (formattingStyle == STYLE_1TBS)
	{
		setBraceFormatMode(LINUX_MODE);
		setAddBracesMode(1);
		setRemoveBracesMode(false);
	}
	else if (formattingStyle == STYLE_GOOGLE)
	{
		setBraceFormatMode(ATTACH_MODE);
		setModifierIndent(true);
		setClassIndent(false);
	}
	else if (formattingStyle == STYLE_MOZILLA)
	{
		setBraceFormatMode(LINUX_MODE);
	}
	else if (formattingStyle == STYLE_WEBKIT)
	{
		setBraceFormatMode(LINUX_MODE);
	}
	else if (formattingStyle == STYLE_PICO)
	{
		setBraceFormatMode(RUN_IN_MODE);
		setAttachClosingBraceMode(true);
		setSwitchIndent(true);
		setBreakOneLineBlocksMode(false);
		setBreakOneLineStatementsMode(false);
		// add-braces won't work for pico, but it could be fixed if necessary
		// both options should be set to true
		if (shouldAddBraces)
			shouldAddOneLineBraces = true;
	}
	else if (formattingStyle == STYLE_LISP)
	{
		setBraceFormatMode(ATTACH_MODE);
		setAttachClosingBraceMode(true);
		setBreakOneLineStatementsMode(false);
		// add-one-line-braces won't work for lisp
		// only shouldAddBraces should be set to true
		if (shouldAddOneLineBraces)
		{
			shouldAddBraces = true;
			shouldAddOneLineBraces = false;
		}
	}
	setMinConditionalIndentLength();
	// if not set by indent=force-tab-x set equal to indentLength
	if (getTabLength() == 0)
		setDefaultTabLength();
	// add-one-line-braces implies keep-one-line-blocks
	if (shouldAddOneLineBraces)
		setBreakOneLineBlocksMode(false);
	// don't allow add-braces and remove-braces
	if (shouldAddBraces || shouldAddOneLineBraces)
		setRemoveBracesMode(false);
	// don't allow break-return-type and attach-return-type
	if (shouldBreakReturnType)
		shouldAttachReturnType = false;
	if (shouldBreakReturnTypeDecl)
		shouldAttachReturnTypeDecl = false;
	// don't allow indent-classes and indent-modifiers
	if (getClassIndent())
		setModifierIndent(false);
}

bool ASFormatter::handleImmediatelyPostHeaderSection()
{
	// should braces be added
	if (currentChar != '{'
	        && shouldAddBraces
	        && currentChar != '#'	// don't add to preprocessor
	        && (shouldBreakOneLineStatements || !isHeaderInMultiStatementLine)
	        && isOkToBreakBlock(braceTypeStack->back()))
	{
		bool bracesAdded = addBracesToStatement();
		if (bracesAdded && !shouldAddOneLineBraces)
		{
			size_t firstText = currentLine.find_first_not_of(" \t");
			assert(firstText != std::string::npos);
			if ((int) firstText == charNum || shouldBreakOneLineHeaders)
				breakCurrentOneLineBlock = true;
		}
	}
	// should braces be removed
	else if (currentChar == '{' && shouldRemoveBraces)
	{
		bool bracesRemoved = removeBracesFromStatement();
		if (bracesRemoved)
		{
			shouldRemoveNextClosingBrace = true;
			if (isBeforeAnyLineEndComment(charNum))
				spacePadNum--;
			else if (shouldBreakOneLineBlocks
			         || (currentLineBeginsWithBrace
			             && currentLine.find_first_not_of(" \t") != std::string::npos))
				shouldBreakLineAtNextChar = true;
			return false;
		}
	}

	// break 'else-if' if shouldBreakElseIfs is requested
	if (shouldBreakElseIfs
	        && currentHeader == &ASResource::AS_ELSE
	        && isOkToBreakBlock(braceTypeStack->back())
	        && !isBeforeAnyComment()
	        && (shouldBreakOneLineStatements || !isHeaderInMultiStatementLine))
	{
		std::string nextText = peekNextText(currentLine.substr(charNum));
		if (!nextText.empty()
		        && isCharPotentialHeader(nextText, 0)
		        && ASBase::findHeader(nextText, 0, headers) == &ASResource::AS_IF)
		{
			isInLineBreak = true;
		}
	}

	// break a header (e.g. if, while, else) from the following statement
	if (shouldBreakOneLineHeaders
	        && peekNextChar() != ' '
	        && (shouldBreakOneLineStatements
	            || (!isHeaderInMultiStatementLine
	                && !isMultiStatementLine()))
	        && isOkToBreakBlock(braceTypeStack->back())
	        && !isBeforeAnyComment())
	{
		if (currentChar == '{')
		{
			if (!currentLineBeginsWithBrace)
			{
				if (isOneLineBlockReached(currentLine, charNum) == 3)
					isInLineBreak = false;
				else
					breakCurrentOneLineBlock = true;
			}
		}
		else if (currentHeader == &ASResource::AS_ELSE)
		{
			std::string nextText = peekNextText(currentLine.substr(charNum), true);
			if (!nextText.empty()
			        && ((isCharPotentialHeader(nextText, 0)
			             && ASBase::findHeader(nextText, 0, headers) != &ASResource::AS_IF)
			            || nextText[0] == '{'))
				isInLineBreak = true;
		}
		else
		{
			// GH16 only break if header is present
			if (currentHeader)
				isInLineBreak = true;
		}
	}

	isImmediatelyPostHeader = false;
	return true;
}

bool ASFormatter::handlePassedSemicolonSection()
{
	isInAllocator = false; // GH16
	isInMultlineStatement = false;
	passedSemicolon = false;

	if (parenStack->back() == 0 && !isCharImmediatelyPostComment && currentChar != ';') // allow ;;
	{
		// does a one-line block have ending comments?
		if (isBraceType(braceTypeStack->back(), SINGLE_LINE_TYPE))
		{
			size_t blockEnd = currentLine.rfind(ASResource::AS_CLOSE_BRACE);
			assert(blockEnd != std::string::npos);
			// move ending comments to this formattedLine
			if (isBeforeAnyLineEndComment(blockEnd))
			{
				size_t commentStart = currentLine.find_first_not_of(" \t", blockEnd + 1);
				assert(commentStart != std::string::npos);
				assert((currentLine.compare(commentStart, 2, "//") == 0)
				       || (currentLine.compare(commentStart, 2, "/*") == 0));
				formattedLine.append(getIndentLength() - 1, ' ');
				// append comment
				int charNumSave = charNum;
				charNum = commentStart;
				while (charNum < (int) currentLine.length())
				{
					currentChar = currentLine[charNum];
					if (currentChar == '\t' && shouldConvertTabs)
						convertTabToSpaces();
					formattedLine.append(1, currentChar);
					++charNum;
				}
				size_t commentLength = currentLine.length() - commentStart;
				currentLine.erase(commentStart, commentLength);
				charNum = charNumSave;
				currentChar = currentLine[charNum];
				testForTimeToSplitFormattedLine();
			}
		}
		isInExecSQL = false;
		shouldReparseCurrentChar = true;
		if (formattedLine.find_first_not_of(" \t") != std::string::npos)
			isInLineBreak = true;
		if (needHeaderOpeningBrace)
		{
			isCharImmediatelyPostCloseBlock = true;
			needHeaderOpeningBrace = false;
		}
		return false;
	}
	return true;
}

void ASFormatter::handleAttachedReturnTypes()
{
	if ((size_t) charNum == methodAttachCharNum)
	{
		int pa = pointerAlignment;
		int ra = referenceAlignment;
		int itemAlignment = (previousNonWSChar == '*' || previousNonWSChar == '^')
		                    ? pa : ((ra == REF_SAME_AS_PTR) ? pa : ra);
		isInLineBreak = false;
		if (previousNonWSChar == '*' || previousNonWSChar == '&' || previousNonWSChar == '^')
		{
			if (itemAlignment == REF_ALIGN_TYPE)
			{
				if (!formattedLine.empty()
				        && !std::isblank(formattedLine[formattedLine.length() - 1]))
					formattedLine.append(1, ' ');
			}
			else if (itemAlignment == REF_ALIGN_MIDDLE)
			{
				if (!formattedLine.empty()
				        && !std::isblank(formattedLine[formattedLine.length() - 1]))
					formattedLine.append(1, ' ');
			}
			else if (itemAlignment == REF_ALIGN_NAME)
			{
				if (!formattedLine.empty()
				        && std::isblank(formattedLine[formattedLine.length() - 1]))
					formattedLine.erase(formattedLine.length() - 1);
			}
			else
			{
				if (formattedLine.length() > 1
				        && !std::isblank(formattedLine[formattedLine.length() - 2]))
					formattedLine.append(1, ' ');
			}
		}
		else
			formattedLine.append(1, ' ');
	}
	methodAttachCharNum = std::string::npos;
	methodAttachLineNum = 0;
}

void ASFormatter::handleClosedBracesOrParens()
{
	foundPreCommandHeader = false;
	parenStack->back()--;
	// this can happen in preprocessor directives
	if (parenStack->back() < 0)
		parenStack->back() = 0;
	if (!questionMarkStack->empty())
	{
		foundQuestionMark = questionMarkStack->back();
		questionMarkStack->pop_back();
	}

	if (isInTemplate && currentChar == '>')
	{
		templateDepth--;
		if (templateDepth == 0)
		{
			isInTemplate = false;
			isImmediatelyPostTemplate = true;
		}
	}

	// check if this parenthesis closes a header, e.g. if (...), while (...)
	//GH16
	if ( !(isSharpStyle() && peekNextChar() == ',') && isInHeader && parenStack->back() == 0)
	{
		isInHeader = false;
		isImmediatelyPostHeader = true;
		foundQuestionMark = false;
	}
	if (currentChar == ']')
	{
		--squareBracketCount;
		if (squareBracketCount <= 0)
		{
			squareBracketCount = 0;
			objCColonAlign = 0;
		}
	}

	// GH16 break
	if (currentChar == ')')
	{
		--parenthesesCount;
		foundCastOperator = false;
		if (parenStack->back() == 0)
			endOfAsmReached = true;
	}
}

void ASFormatter::handleBraces()
{
	// if appendOpeningBrace this was already done for the original brace
	if (currentChar == '{' && !appendOpeningBrace)
	{
		BraceType newBraceType = getBraceType();
		breakCurrentOneLineBlock = false;
		foundNamespaceHeader = false;
		foundClassHeader = false;
		foundStructHeader = false;
		foundInterfaceHeader = false;
		foundPreDefinitionHeader = false;
		foundPreCommandHeader = false;
		foundPreCommandMacro = false;
		foundTrailingReturnType = false;
		isInPotentialCalculation = false;
		isInObjCMethodDefinition = false;
		isImmediatelyPostObjCMethodPrefix = false;
		isInObjCInterface = false;
		isInEnum = false;

		isJavaStaticConstructor = false;
		isCharImmediatelyPostNonInStmt = false;
		needHeaderOpeningBrace = false;
		shouldKeepLineUnbroken = false;
		returnTypeChecked = false;

		isInExplicitBlock++;

		objCColonAlign = 0;

		methodBreakCharNum = std::string::npos;
		methodBreakLineNum = 0;
		methodAttachCharNum = std::string::npos;
		methodAttachLineNum = 0;

		isPreviousBraceBlockRelated = !isBraceType(newBraceType, ARRAY_TYPE);
		braceTypeStack->emplace_back(newBraceType);
		preBraceHeaderStack->emplace_back(currentHeader);
		currentHeader = nullptr;
		// do not use emplace_back on std::vector<bool> until supported by macOS
		structStack->push_back(isInIndentableStruct);
		if (isBraceType(newBraceType, STRUCT_TYPE) && isCStyle())
			isInIndentableStruct = isStructAccessModified(currentLine, charNum);
		else
			isInIndentableStruct = false;

		bracesNestingLevel++;
	}

	// this must be done before the braceTypeStack is popped
	BraceType braceType = braceTypeStack->back();
	bool isOpeningArrayBrace = (isBraceType(braceType, ARRAY_TYPE)
	                            && braceTypeStack->size() >= 2
	                            && !isBraceType((*braceTypeStack)[braceTypeStack->size() - 2], ARRAY_TYPE)
	                           );

	if (currentChar == '}')
	{
		// if a request has been made to append a post block empty line,
		// but the block exists immediately before a closing brace,
		// then there is no need for the post block empty line.
		isAppendPostBlockEmptyLineRequested = false;
		if (isInAsm)
			endOfAsmReached = true;
		isInAsmOneLine = isInQuote = false;
		shouldKeepLineUnbroken = false;
		squareBracketCount = 0;
		isInAllocator = false;
		isInMultlineStatement = false;
		isInExplicitBlock--;

		if (braceTypeStack->size() > 1)
		{
			previousBraceType = braceTypeStack->back();
			braceTypeStack->pop_back();
			isPreviousBraceBlockRelated = !isBraceType(braceType, ARRAY_TYPE);
		}
		else
		{
			previousBraceType = NULL_TYPE;
			isPreviousBraceBlockRelated = false;
		}

		if (!preBraceHeaderStack->empty())
		{
			previousHeader = currentHeader;
			currentHeader = preBraceHeaderStack->back();
			preBraceHeaderStack->pop_back();
		}
		else
			currentHeader = nullptr;

		if (!structStack->empty())
		{
			isInIndentableStruct = structStack->back();
			structStack->pop_back();
		}
		else
			isInIndentableStruct = false;

		if (isNonInStatementArray
		        && (!isBraceType(braceTypeStack->back(), ARRAY_TYPE)	// check previous brace
		            || peekNextChar() == ';'))							// check for "};" added V2.01
			isImmediatelyPostNonInStmt = true;

		if (!shouldBreakOneLineStatements
		        && ASBeautifier::getNextWord(currentLine, charNum) == ASResource::AS_ELSE)
		{
			// handle special case of "else" at the end of line
			size_t nextText = currentLine.find_first_not_of(" \t", charNum + 1);
			if (ASBeautifier::peekNextChar(currentLine, nextText + 3) == ' ')
				shouldBreakLineAtNextChar = true;
		}
		bracesNestingLevel--;
	}

	// format braces
	appendOpeningBrace = false;
	if (isBraceType(braceType, ARRAY_TYPE))
	{
		formatArrayBraces(braceType, isOpeningArrayBrace);
	}
	else
	{
		if (currentChar == '{')
			formatOpeningBrace(braceType);
		else
			formatClosingBrace(braceType);
	}
}


void ASFormatter::handleBreakLine()
{
	isCharImmediatelyPostOpenBlock = (previousCommandChar == '{');
	isCharImmediatelyPostCloseBlock = (previousCommandChar == '}');

	if (isCharImmediatelyPostOpenBlock
	        && !isCharImmediatelyPostComment
	        && !isCharImmediatelyPostLineComment)
	{
		previousCommandChar = ' ';

		if (braceFormatMode == NONE_MODE)
		{
			if (isBraceType(braceTypeStack->back(), SINGLE_LINE_TYPE)
			        && (isBraceType(braceTypeStack->back(), BREAK_BLOCK_TYPE)
			            || shouldBreakOneLineBlocks))
				isInLineBreak = true;
			else if (currentLineBeginsWithBrace)
				formatRunIn();
			else
				breakLine();
		}
		else if (braceFormatMode == RUN_IN_MODE
		         && currentChar != '#')
			formatRunIn();
		else
			isInLineBreak = true;
	}
	else if (isCharImmediatelyPostCloseBlock
	         && shouldBreakOneLineStatements
	         && !isCharImmediatelyPostComment
	         && ((isLegalNameChar(currentChar) && currentChar != '.')
	             || currentChar == '+'
	             || currentChar == '-'
	             || currentChar == '*'
	             || currentChar == '&'
	             || currentChar == '('))
	{
		previousCommandChar = ' ';
		isInLineBreak = true;
	}
}

bool ASFormatter::handlePotentialHeader(const std::string* newHeader)
{
	isNonParenHeader = false;
	foundClosingHeader = false;

	newHeader = findHeader(headers);

	// java can have a 'default' not in a switch
	if (newHeader == &ASResource::AS_DEFAULT
	        && ASBeautifier::peekNextChar(
	            currentLine, charNum + (*newHeader).length() - 1) != ':')
		newHeader = nullptr;
	// Qt headers may be variables in C++
	if (isCStyle()
	        && (newHeader == &ASResource::AS_FOREVER || newHeader == &ASResource::AS_FOREACH))
	{
		if (currentLine.find_first_of("=;", charNum) != std::string::npos)
			newHeader = nullptr;
	}
	if (isJavaStyle()
	        && (newHeader == &ASResource::AS_SYNCHRONIZED))
	{
		// want synchronized statements not synchronized methods
		if (!isBraceType(braceTypeStack->back(), COMMAND_TYPE))
			newHeader = nullptr;
	}
	else if (newHeader == &ASResource::AS_USING
	         && ASBeautifier::peekNextChar(
	             currentLine, charNum + (*newHeader).length() - 1) != '(')
		newHeader = nullptr;

	if (newHeader != nullptr)
	{
		foundClosingHeader = isClosingHeader(newHeader);

		if (!foundClosingHeader)
		{
			// these are closing headers
			if ((newHeader == &ASResource::AS_WHILE && currentHeader == &ASResource::AS_DO)
			        || (newHeader == &ASResource::_AS_FINALLY && currentHeader == &ASResource::_AS_TRY)
			        || (newHeader == &ASResource::_AS_EXCEPT && currentHeader == &ASResource::_AS_TRY))
				foundClosingHeader = true;
			// don't append empty block for these related headers
			else if (isSharpStyle()
			         && previousNonWSChar == '}'
			         && ((newHeader == &ASResource::AS_SET && currentHeader == &ASResource::AS_GET)
			             || (newHeader == &ASResource::AS_REMOVE && currentHeader == &ASResource::AS_ADD))
			         && isOkToBreakBlock(braceTypeStack->back()))
				isAppendPostBlockEmptyLineRequested = false;
		}

		previousHeader = currentHeader;
		currentHeader = newHeader;
		needHeaderOpeningBrace = true;

		// is the previous statement on the same line?
		if ((previousNonWSChar == ';' || previousNonWSChar == ':')
		        && !isInLineBreak
		        && isOkToBreakBlock(braceTypeStack->back()))
		{
			// if breaking lines, break the line at the header
			// except for multiple 'case' statements on a line
			if (maxCodeLength != std::string::npos
			        && previousHeader != &ASResource::AS_CASE)
				isInLineBreak = true;
			else
				isHeaderInMultiStatementLine = true;
		}

		if (foundClosingHeader && previousNonWSChar == '}')
		{
			if (isOkToBreakBlock(braceTypeStack->back()))
				isLineBreakBeforeClosingHeader();

			// get the adjustment for a comment following the closing header
			if (isInLineBreak)
				nextLineSpacePadNum = getNextLineCommentAdjustment();
			else
				spacePadNum = getCurrentLineCommentAdjustment();
		}

		// check if the found header is non-paren header
		isNonParenHeader = findHeader(nonParenHeaders) != nullptr;

		if (isNonParenHeader
		        && (currentHeader == &ASResource::AS_CATCH
		            || currentHeader == &ASResource::AS_CASE))
		{
			int startChar = charNum + currentHeader->length() - 1;
			if (ASBeautifier::peekNextChar(currentLine, startChar) == '(')
				isNonParenHeader = false;
		}

		// join 'else if' statements
		if (currentHeader == &ASResource::AS_IF
		        && previousHeader == &ASResource::AS_ELSE
		        && isInLineBreak
		        && !shouldBreakElseIfs
		        && !isCharImmediatelyPostLineComment
		        && !isImmediatelyPostPreprocessor)
		{
			// 'else' must be last thing on the line
			size_t start = formattedLine.length() >= 6 ? formattedLine.length() - 6 : 0;
			if (formattedLine.find(ASResource::AS_ELSE, start) != std::string::npos)
			{
				appendSpacePad();
				isInLineBreak = false;
			}
		}

		appendSequence(*currentHeader);
		goForward(currentHeader->length() - 1);
		// if a paren-header is found add a space after it, if needed
		// this checks currentLine, appendSpacePad() checks formattedLine
		// in 'case' and C# 'catch' can be either a paren or non-paren header
		if (shouldPadHeader
		        && !isNonParenHeader
		        && charNum < (int) currentLine.length() - 1 && !std::isblank(currentLine[charNum + 1]))
			appendSpacePad();

		// Signal that a header has been reached
		// *** But treat a closing while() (as in do...while)
		//     as if it were NOT a header since a closing while()
		//     should never have a block after it!
		if (currentHeader != &ASResource::AS_CASE && currentHeader != &ASResource::AS_DEFAULT
		        && !(foundClosingHeader && currentHeader == &ASResource::AS_WHILE))
		{
			isInHeader = true;

			// in C# 'catch' and 'delegate' can be a paren or non-paren header
			if (isNonParenHeader && !isSharpStyleWithParen(currentHeader))
			{
				isImmediatelyPostHeader = true;
				isInHeader = false;
			}
		}

		// #569
		if (shouldBreakBlocks
		        && isOkToBreakBlock(braceTypeStack->back())
		        && !isHeaderInMultiStatementLine)
		{
			if (previousHeader == nullptr
			        && !foundClosingHeader
			        && !isCharImmediatelyPostOpenBlock
			        && !isImmediatelyPostCommentOnly)
			{
				isPrependPostBlockEmptyLineRequested = true;
			}

			if (isClosingHeader(currentHeader)
			        || foundClosingHeader)
			{
				isPrependPostBlockEmptyLineRequested = false;
			}

			if (shouldBreakClosingHeaderBlocks
			        && isCharImmediatelyPostCloseBlock
			        && !isImmediatelyPostCommentOnly
			        && !(currentHeader == &ASResource::AS_WHILE			// do-while
			             && foundClosingHeader))
			{
				isPrependPostBlockEmptyLineRequested = true;
			}
		}

		if (currentHeader == &ASResource::AS_CASE
		        || currentHeader == &ASResource::AS_DEFAULT)
			isInCase = true;

		return false;
	}
	if ((newHeader = findHeader(preDefinitionHeaders)) != nullptr
	        && parenStack->back() == 0
	        && !isInEnum)		// not C++11 enum class
	{
		if (newHeader == &ASResource::AS_NAMESPACE || newHeader == &ASResource::AS_MODULE)
			foundNamespaceHeader = true;
		if (newHeader == &ASResource::AS_CLASS)
			foundClassHeader = true;
		if (newHeader == &ASResource::AS_STRUCT)
			foundStructHeader = true;
		if (newHeader == &ASResource::AS_INTERFACE && !foundNamespaceHeader && !foundClassHeader)
			foundInterfaceHeader = true;
		foundPreDefinitionHeader = true;
		appendSequence(*newHeader);
		goForward(newHeader->length() - 1);

		return false;
	}
	if ((newHeader = findHeader(preCommandHeaders)) != nullptr)
	{
		// must be after function arguments
		if (previousNonWSChar == ')')
			foundPreCommandHeader = true;
	}
	else if ((newHeader = findHeader(castOperators)) != nullptr)
	{
		foundCastOperator = true;
		appendSequence(*newHeader);
		goForward(newHeader->length() - 1);
		return false;
	}
	return true;
}


void ASFormatter::handleEndOfBlock()
{
	if (currentChar == ';' && !isInAsmBlock)
	{
		squareBracketCount = 0;

		methodBreakCharNum = std::string::npos;
		methodBreakLineNum = 0;
		methodAttachCharNum = std::string::npos;
		methodAttachLineNum = 0;

		if (((shouldBreakOneLineStatements
		        || isBraceType(braceTypeStack->back(), SINGLE_LINE_TYPE))
		        && isOkToBreakBlock(braceTypeStack->back()))
		        && !(attachClosingBraceMode && peekNextChar() == '}'))
		{
			passedSemicolon = true;
		}
		else if (!shouldBreakOneLineStatements
		         && ASBeautifier::getNextWord(currentLine, charNum) == ASResource::AS_ELSE)
		{
			// handle special case of "else" at the end of line
			size_t nextText = currentLine.find_first_not_of(" \t", charNum + 1);
			if (ASBeautifier::peekNextChar(currentLine, nextText + 3) == ' ')
				passedSemicolon = true;
		}

//is set in struct case? #518 #569
		if (shouldBreakBlocks
		        && currentHeader != nullptr
		        && currentHeader != &ASResource::AS_CASE
		        && currentHeader != &ASResource::AS_DEFAULT
		        && !isHeaderInMultiStatementLine
		        && parenStack->back() == 0
		   )
		{
			isAppendPostBlockEmptyLineRequested = true;
		}
	}
	if (currentChar != ';'
	        || foundStructHeader // #518
	        || (needHeaderOpeningBrace && parenStack->back() == 0))
		currentHeader = nullptr;

	resetEndOfStatement();
}

void ASFormatter::handleColonSection()
{
	if (isInCase)
	{
		isInCase = false;
		if (shouldBreakOneLineStatements)
			passedColon = true;
	}
	else if (isCStyle()                     // for C/C++ only
	         && isOkToBreakBlock(braceTypeStack->back())
	         && shouldBreakOneLineStatements
	         && !foundQuestionMark          // not in a ?: sequence
	         && !foundPreDefinitionHeader   // not in a definition block
	         && previousCommandChar != ')'  // not after closing paren of a method header
	         && !foundPreCommandHeader      // not after a 'noexcept'
	         && squareBracketCount == 0     // not in objC method call
	         && !isInObjCMethodDefinition   // not objC '-' or '+' method
	         && !isInObjCInterface          // not objC @interface
	         && !isInObjCSelector           // not objC @selector
	         && !isDigit(peekNextChar()) && !lineStartsWithNumericType(currentLine)    // not a bit field xxxx
	         && !isInEnum                   // not an enum with a base type
	         && !isInStruct                 // not an struct
	         && !isInContinuedPreProc           // not in preprocessor line
	         && !isInAsm                    // not in extended assembler
	         && !isInAsmOneLine             // not in extended assembler
	         && !isInAsmBlock)              // not in extended assembler
	{
		passedColon = true;
	}

	if (isObjCStyle()
	        && (squareBracketCount > 0 || isInObjCMethodDefinition || isInObjCSelector)
	        && !foundQuestionMark)			// not in a ?: sequence
	{
		isImmediatelyPostObjCMethodPrefix = false;
		isInObjCReturnType = false;
		isInObjCParam = true;
		if (shouldPadMethodColon)
			padObjCMethodColon();
	}

	if (isInObjCInterface)
	{
		appendSpacePad();
		if ((int) currentLine.length() > charNum + 1
		        && !std::isblank(currentLine[charNum + 1]))
			currentLine.insert(charNum + 1, " ");
	}

	if (isClassInitializer())
	{
		isInClassInitializer = true;
	}
}

void ASFormatter::handlePotentialHeaderPart2()
{
	//GL30
	if (!isGSCStyle() && (findKeyword(currentLine, charNum, ASResource::AS_NEW)
	                      || findKeyword(currentLine, charNum, ASResource::AS_DELETE)))
	{
		isInPotentialCalculation = false;
		isImmediatelyPostNewDelete = true;
	}

	//https://sourceforge.net/p/astyle/bugs/464/ + GH16
	if (isSharpStyle() && findKeyword(currentLine, charNum, ASResource::AS_NEW)
	        && currentHeader != &ASResource::AS_FOREACH
	        && currentHeader != &ASResource::AS_FOR
	        && currentHeader != &ASResource::AS_USING
	        && currentHeader != &ASResource::AS_WHILE
	        && currentHeader != &ASResource::AS_IF
	        && currentLine.find(ASResource::AS_PUBLIC) == std::string::npos
	        && currentLine.find(ASResource::AS_PROTECTED) == std::string::npos
	        && currentLine.find(ASResource::AS_PRIVATE) == std::string::npos
	   )
	{
		isInAllocator = true;
	}

	if (findKeyword(currentLine, charNum, ASResource::AS_RETURN))
	{
		isInPotentialCalculation = true;
		isImmediatelyPostReturn = true;		// return is the same as an = sign
	}

	if (findKeyword(currentLine, charNum, ASResource::AS_OPERATOR))
		isImmediatelyPostOperator = true;

	if (findKeyword(currentLine, charNum, ASResource::AS_ENUM))
	{
		size_t firstNum = currentLine.find_first_of("(){},/");
		if (firstNum == std::string::npos
		        || currentLine[firstNum] == '{'
		        || currentLine[firstNum] == '/')
			isInEnum = true;
	}

	if (findKeyword(currentLine, charNum, ASResource::AS_TYPEDEF_STRUCT) || findKeyword(currentLine, charNum, ASResource::AS_STRUCT))
	{
		size_t firstNum = currentLine.find_first_of("(){},/");

		if (firstNum == std::string::npos
		        || currentLine[firstNum] == '{'
		        || currentLine[firstNum] == '/')
		{
			isInStruct = true;
		}
	}

	if (isCStyle()
	        && findKeyword(currentLine, charNum, ASResource::AS_THROW)
	        && previousCommandChar != ')'
	        && !foundPreCommandHeader)      // 'const' throw()
		isImmediatelyPostThrow = true;

	if (isCStyle() && findKeyword(currentLine, charNum, ASResource::AS_EXTERN) && isExternC())
		isInExternC = true;

	if (isCStyle() && findKeyword(currentLine, charNum, ASResource::AS_AUTO)
	        && (isBraceType(braceTypeStack->back(), NULL_TYPE)
	            || isBraceType(braceTypeStack->back(), DEFINITION_TYPE))
	        && (currentLine.find('(') != std::string::npos)) // #516 auto array initializer with braces should not be blocks
		foundTrailingReturnType = true;

	// check for break/attach return type
	if (shouldBreakReturnType || shouldBreakReturnTypeDecl
	        || shouldAttachReturnType || shouldAttachReturnTypeDecl)
	{
		if ((isBraceType(braceTypeStack->back(), NULL_TYPE)
		        || isBraceType(braceTypeStack->back(), DEFINITION_TYPE))
		        && !returnTypeChecked
		        && !foundNamespaceHeader
		        && !foundClassHeader
		        && !isInObjCMethodDefinition
		        // bypass objective-C and java @ character
		        && charNum == (int) currentLine.find_first_not_of(" \t")

		        // possibly related to #504
		        && !(isCStyle() && isCharPotentialHeader(currentLine, charNum)
		             && (findKeyword(currentLine, charNum, ASResource::AS_PUBLIC)
		                 || findKeyword(currentLine, charNum, ASResource::AS_PRIVATE)
		                 || findKeyword(currentLine, charNum, ASResource::AS_PROTECTED)))
		   )
		{
			findReturnTypeSplitPoint(currentLine);
			returnTypeChecked = true;
		}
	}

	// Objective-C NSException macros are preCommandHeaders
	if (isCStyle() && findKeyword(currentLine, charNum, ASResource::AS_NS_DURING))
		foundPreCommandMacro = true;
	if (isCStyle() && findKeyword(currentLine, charNum, ASResource::AS_NS_HANDLER))
		foundPreCommandMacro = true;

	if (isCStyle() && isExecSQL(currentLine, charNum))
		isInExecSQL = true;

	if (isCStyle())
	{
		if (findKeyword(currentLine, charNum, ASResource::AS_ASM)
		        || findKeyword(currentLine, charNum, ASResource::AS__ASM__))
		{
			isInAsm = true;
		}
		else if (findKeyword(currentLine, charNum, ASResource::AS_MS_ASM)		// microsoft specific
		         || findKeyword(currentLine, charNum, ASResource::AS_MS__ASM))
		{
			int index = 4;
			if (peekNextChar() == '_')	// check for __asm
				index = 5;

			char peekedChar = ASBase::peekNextChar(currentLine, charNum + index);
			if (peekedChar == '{' || peekedChar == ' ')
				isInAsmBlock = true;
			else
				isInAsmOneLine = true;
		}
	}

	if (isJavaStyle()
	        && (findKeyword(currentLine, charNum, ASResource::AS_STATIC)
	            && isNextCharOpeningBrace(charNum + 6)))
		isJavaStaticConstructor = true;

	if (isSharpStyle()
	        && (findKeyword(currentLine, charNum, ASResource::AS_DELEGATE)
	            || findKeyword(currentLine, charNum, ASResource::AS_UNCHECKED)))
		isSharpDelegate = true;

	// append the entire name
	std::string_view name = getCurrentWord(currentLine, charNum);
	// must pad the 'and' and 'or' operators if required
	if (name == "and" || name == "or")
	{
		if (shouldPadOperators && previousNonWSChar != ':')
		{
			appendSpacePad();
			appendOperator(std::string(name));
			goForward(name.length() - 1);
			if (!isBeforeAnyComment()
			        && !(currentLine.compare(charNum + 1, 1, ASResource::AS_SEMICOLON) == 0)
			        && !(currentLine.compare(charNum + 1, 2, ASResource::AS_SCOPE_RESOLUTION) == 0))
				appendSpaceAfter();
		}
		else
		{
			appendOperator(std::string(name));
			goForward(name.length() - 1);
		}
	}
	else
	{
		appendSequence(std::string(name));
		goForward(name.length() - 1);
	}
}

void ASFormatter::handlePotentialOperator(const std::string *newHeader)
{

	// check for Java ? wildcard
	if (newHeader != nullptr
	        && newHeader == &ASResource::AS_GCC_MIN_ASSIGN
	        && isJavaStyle()
	        && isInTemplate)
		newHeader = nullptr;

	if (newHeader != nullptr)
	{
		if (newHeader == &ASResource::AS_LAMBDA)
			foundPreCommandHeader = true;

		// correct mistake of two >> closing a template
		if (isInTemplate && (newHeader == &ASResource::AS_GR_GR || newHeader == &ASResource::AS_GR_GR_GR))
			newHeader = &ASResource::AS_GR;

		if (!isInPotentialCalculation)
		{
			// must determine if newHeader is an assignment operator
			// do NOT use findOperator - the length must be exact!!!
			if (find(begin(*assignmentOperators), end(*assignmentOperators), newHeader)
			        != end(*assignmentOperators))
			{
				foundPreCommandHeader = false;
				char peekedChar = peekNextChar();
				isInPotentialCalculation = !(newHeader == &ASResource::AS_EQUAL && peekedChar == '*')
				                           && !(newHeader == &ASResource::AS_EQUAL && peekedChar == '&')
				                           && !isCharImmediatelyPostOperator;
			}
		}
	}
}

void ASFormatter::handleParens()
{
	if (currentChar == '(')
	{
		if (shouldPadHeader
		        && (isCharImmediatelyPostReturn
		            || isCharImmediatelyPostThrow
		            || isCharImmediatelyPostNewDelete))
			appendSpacePad();
	}

	if (shouldPadParensOutside || shouldPadParensInside || shouldUnPadParens || shouldPadFirstParen)
		padParensOrBrackets('(', ')', shouldPadFirstParen);
	else
		appendCurrentChar();

	if (isInObjCMethodDefinition)
	{
		if (currentChar == '(' && isImmediatelyPostObjCMethodPrefix)
		{
			if (shouldPadMethodPrefix || shouldUnPadMethodPrefix)
				padObjCMethodPrefix();
			isImmediatelyPostObjCMethodPrefix = false;
			isInObjCReturnType = true;
		}
		else if (currentChar == ')' && isInObjCReturnType)
		{
			if (shouldPadReturnType || shouldUnPadReturnType)
				padObjCReturnType();
			isInObjCReturnType = false;
		}
		else if (isInObjCParam
		         && (shouldPadParamType || shouldUnPadParamType))
			padObjCParamType();
	}
}


void ASFormatter::handleOpenParens()
{
	questionMarkStack->push_back(foundQuestionMark);
	foundQuestionMark = false;
	parenStack->back()++;
	if (currentChar == '[')
	{
		++squareBracketCount;
		if (getAlignMethodColon() && squareBracketCount == 1 &&
		        isCStyle())
			objCColonAlign = findObjCColonAlignment();
	}
	if (currentChar == '(')
	{
		++parenthesesCount;
	}
}

void ASFormatter::formatFirstOpenBrace(BraceType braceType)
{
	if (braceFormatMode == ATTACH_MODE || braceFormatMode == LINUX_MODE)
	{
		// break an enum if mozilla
		if (isBraceType(braceType, ENUM_TYPE)
		        && formattingStyle == STYLE_MOZILLA
		        && !(!shouldBreakOneLineBlocks && formattedLine.find('}' != std::string::npos) ) // GL38
		   )
		{
			isInLineBreak = true;
			appendCurrentChar();                // don't attach
		}
		// don't attach to a preprocessor directive or '\' line
		else if ((isImmediatelyPostPreprocessor
		          || (!formattedLine.empty()
		              && formattedLine[formattedLine.length() - 1] == '\\'))
		         && currentLineBeginsWithBrace)
		{
			isInLineBreak = true;
			appendCurrentChar();                // don't attach
		}
		else if (isCharImmediatelyPostComment)
		{
			// TODO: attach brace to line-end comment
			appendCurrentChar();                // don't attach
		}
		else if (isCharImmediatelyPostLineComment && !isBraceType(braceType, SINGLE_LINE_TYPE))
		{
			appendCharInsideComments();
		}
		else
		{
			// if a blank line precedes this don't attach
			if (isEmptyLine(formattedLine))
				appendCurrentChar();            // don't attach
			else
			{
				// if brace is broken or not an assignment
				if (currentLineBeginsWithBrace
				        && !isBraceType(braceType, SINGLE_LINE_TYPE))
				{
					appendSpacePad();
					appendCurrentChar(false);				// OK to attach
					// TODO: debug the following line
					testForTimeToSplitFormattedLine();		// line length will have changed

					if (currentLineBeginsWithBrace && currentLineFirstBraceNum == (size_t) charNum)
						shouldBreakLineAtNextChar = true;
				}
				else
				{
					if (previousNonWSChar != '(')
					{
						// don't space pad C++11 uniform initialization
						if (!isBraceType(braceType, INIT_TYPE))
							appendSpacePad();
					}
					appendCurrentChar();
				}
			}
		}
	}
	else if (braceFormatMode == BREAK_MODE)
	{
		if (std::isblank(peekNextChar()) && !isInVirginLine)
			breakLine();
		else if (isBeforeAnyComment() && sourceIterator->hasMoreLines())
		{
			// do not break unless comment is at line end
			if (isBeforeAnyLineEndComment(charNum) && !currentLineBeginsWithBrace)
			{
				currentChar = ' ';            // remove brace from current line
				appendOpeningBrace = true;    // append brace to following line
			}
		}
		if (!isInLineBreak && previousNonWSChar != '(')
		{
			// don't space pad C++11 uniform initialization
			if (!isBraceType(braceType, INIT_TYPE))
				appendSpacePad();
		}
		appendCurrentChar();

		if (currentLineBeginsWithBrace
		        && currentLineFirstBraceNum == (size_t) charNum
		        && !isBraceType(braceType, SINGLE_LINE_TYPE))
			shouldBreakLineAtNextChar = true;
	}
	else if (braceFormatMode == RUN_IN_MODE)
	{
		if (std::isblank(peekNextChar()) && !isInVirginLine)
			breakLine();
		else if (isBeforeAnyComment() && sourceIterator->hasMoreLines())
		{
			// do not break unless comment is at line end
			if (isBeforeAnyLineEndComment(charNum) && !currentLineBeginsWithBrace)
			{
				currentChar = ' ';            // remove brace from current line
				appendOpeningBrace = true;    // append brace to following line
			}
		}
		if (!isInLineBreak && previousNonWSChar != '(')
		{
			// don't space pad C++11 uniform initialization
			if (!isBraceType(braceType, INIT_TYPE))
				appendSpacePad();
		}
		appendCurrentChar();
	}
	else if (braceFormatMode == NONE_MODE)
	{
		if (currentLineBeginsWithBrace
		        && (size_t) charNum == currentLineFirstBraceNum)
		{
			appendCurrentChar();                // don't attach
		}
		else
		{
			if (previousNonWSChar != '(')
			{
				// don't space pad C++11 uniform initialization
				if (!isBraceType(braceType, INIT_TYPE))
					appendSpacePad();
			}
			appendCurrentChar(false);           // OK to attach
		}
	}
}

void ASFormatter::formatOpenBrace()
{
	if (braceFormatMode == RUN_IN_MODE)
	{
		if (previousNonWSChar == '{'
		        && braceTypeStack->size() > 2
		        && !isBraceType((*braceTypeStack)[braceTypeStack->size() - 2],
		                        SINGLE_LINE_TYPE))
			formatArrayRunIn();
	}
	else if (!isInLineBreak
	         && !std::isblank(peekNextChar())
	         && previousNonWSChar == '{'
	         && braceTypeStack->size() > 2
	         && !isBraceType((*braceTypeStack)[braceTypeStack->size() - 2],
	                         SINGLE_LINE_TYPE))
		formatArrayRunIn();

	appendCurrentChar();
}

void ASFormatter::formatCloseBrace(BraceType braceType)
{
	if (attachClosingBraceMode)
	{
		if (isEmptyLine(formattedLine)			// if a blank line precedes this
		        || isImmediatelyPostPreprocessor
		        || isCharImmediatelyPostLineComment
		        || isCharImmediatelyPostComment)
			appendCurrentChar();				// don't attach
		else
		{
			appendSpacePad();
			appendCurrentChar(false);			// attach
		}
	}
	else
	{
		// does this close the first opening brace in the array?
		// must check if the block is still a single line because of anonymous statements
		if (!isBraceType(braceType, INIT_TYPE)
		        && (!isBraceType(braceType, SINGLE_LINE_TYPE)
		            || formattedLine.find('{') == std::string::npos))
			breakLine();
		appendCurrentChar();
	}

	// if a declaration follows an enum definition, space pad
	char peekedChar = peekNextChar();
	if ((isLegalNameChar(peekedChar) && peekedChar != '.')
	        || peekedChar == '[')
		appendSpaceAfter();
}



std::string ASFormatter::nextLine()
{
	const std::string* newHeader = nullptr;
	isInVirginLine = isVirgin;
	isCharImmediatelyPostComment = false;
	isPreviousCharPostComment = false;
	isCharImmediatelyPostLineComment = false;
	isCharImmediatelyPostOpenBlock = false;
	isCharImmediatelyPostCloseBlock = false;
	isCharImmediatelyPostTemplate = false;

	while (!isLineReady)
	{
		if (shouldReparseCurrentChar)
			shouldReparseCurrentChar = false;
		else if (!getNextChar())
		{
			breakLine();
			continue;
		}
		else // stuff to do when reading a new character...
		{
			// make sure that a virgin '{' at the beginning of the file will be treated as a block...
			if (isInVirginLine && currentChar == '{'
			        && currentLineBeginsWithBrace
			        && previousCommandChar == ' ')
				previousCommandChar = '{';
			if (isInClassInitializer
			        && isBraceType(braceTypeStack->back(), COMMAND_TYPE))
				isInClassInitializer = false;
			if (isInBraceRunIn)
				isInLineBreak = false;
			if (!std::isblank(currentChar))
				isInBraceRunIn = false;
			isPreviousCharPostComment = isCharImmediatelyPostComment;
			isCharImmediatelyPostComment = false;
			isCharImmediatelyPostTemplate = false;
			isCharImmediatelyPostReturn = false;
			isCharImmediatelyPostThrow = false;
			isCharImmediatelyPostNewDelete = false;
			isCharImmediatelyPostOperator = false;
			isCharImmediatelyPostPointerOrReference = false;
			isCharImmediatelyPostOpenBlock = false;
			isCharImmediatelyPostCloseBlock = false;
		}

		if ((lineIsLineCommentOnly || lineIsCommentOnly)
		        && currentLine.find("*INDENT-ON*", charNum) != std::string::npos
		        && isFormattingModeOff)
		{
			isFormattingModeOff = false;
			breakLine();
			formattedLine = currentLine;
			charNum = (int) currentLine.length() - 1;
			continue;
		}
		if (isFormattingModeOff)
		{
			breakLine();
			formattedLine = currentLine;
			charNum = (int) currentLine.length() - 1;
			continue;
		}

		if ((lineIsLineCommentOnly || lineIsCommentOnly)
		        && currentLine.find("*INDENT-OFF*", charNum) != std::string::npos)
		{
			isFormattingModeOff = true;
			if (isInLineBreak)			// is true if not the first line
				breakLine();
			formattedLine = currentLine;
			charNum = (int) currentLine.length() - 1;
			continue;
		}

		if (shouldBreakLineAtNextChar)
		{
			if (std::isblank(currentChar) && !lineIsEmpty)
				continue;
			isInLineBreak = true;
			shouldBreakLineAtNextChar = false;
		}

		if (isInExecSQL && !passedSemicolon)
		{
			if (currentChar == ';')
				passedSemicolon = true;
			appendCurrentChar();
			continue;
		}

		if (isInLineComment)
		{
			formatLineCommentBody();
			continue;
		}

		if (isInComment)
		{
			formatCommentBody();
			continue;
		}

		if (isInQuote)
		{
			formatQuoteBody();
			continue;
		}

		// not in quote or comment or line comment

		if (isSequenceReached(ASResource::AS_OPEN_LINE_COMMENT))
		{
			formatLineCommentOpener();
			testForTimeToSplitFormattedLine();
			continue;
		}
		if (isSequenceReached(ASResource::AS_OPEN_COMMENT) || (isGSCStyle() && isSequenceReached(ASResource::AS_GSC_OPEN_COMMENT)))
		{
			formatCommentOpener();
			testForTimeToSplitFormattedLine();
			continue;
		}
		if (currentChar == '"'
		        || (currentChar == '\'' && !isDigitSeparator(currentLine, charNum)))
		{
			formatQuoteOpener();
			testForTimeToSplitFormattedLine();
			continue;
		}
		// treat these preprocessor statements as a line comment

		if ((currentChar == '#')
		        && currentLine.find_first_not_of(" \t") == (size_t) charNum)
		{
			isInContinuedPreProc = currentLine[currentLine.size() - 1] == '\\';
		}

		if (isInPreprocessor)
		{
			appendCurrentChar();
			continue;
		}

		if (isInTemplate && shouldCloseTemplates)
		{
			if (previousNonWSChar == '>' && std::isblank(currentChar) && peekNextChar() == '>')
				continue;
		}

		if (shouldRemoveNextClosingBrace && currentChar == '}')
		{
			currentLine[charNum] = currentChar = ' ';
			shouldRemoveNextClosingBrace = false;
			assert(adjustChecksumIn(-'}'));
			if (isEmptyLine(currentLine))
				continue;
		}

		// handle white space - needed to simplify the rest.
		if (std::isblank(currentChar))
		{
			appendCurrentChar();
			continue;
		}

		/* not in MIDDLE of quote or comment or SQL or white-space of any type ... */

		// check if in preprocessor
		// ** isInPreprocessor will be automatically reset at the beginning
		//    of a new line in getnextChar()
		if (currentChar == '#' && !isBraceType(braceTypeStack->back(), SINGLE_LINE_TYPE))
		{
			isInPreprocessor = true;
			// check for run-in
			if (!formattedLine.empty() && formattedLine[0] == '{')
			{
				isInLineBreak = true;
				isInBraceRunIn = false;
			}
			processPreprocessor();
		}

		/* not in preprocessor ... */

		if (isImmediatelyPostComment)
		{
			caseHeaderFollowsComments = false;
			isImmediatelyPostComment = false;
			isCharImmediatelyPostComment = true;
		}

		if (isImmediatelyPostLineComment)
		{
			caseHeaderFollowsComments = false;
			isImmediatelyPostLineComment = false;
			isCharImmediatelyPostLineComment = true;
		}

		if (isImmediatelyPostReturn)
		{
			isImmediatelyPostReturn = false;
			isCharImmediatelyPostReturn = true;
		}

		if (isImmediatelyPostThrow)
		{
			isImmediatelyPostThrow = false;
			isCharImmediatelyPostThrow = true;
		}

		if (isImmediatelyPostNewDelete)
		{
			isImmediatelyPostNewDelete = false;
			isCharImmediatelyPostNewDelete = true;
		}

		if (isImmediatelyPostOperator)
		{
			isImmediatelyPostOperator = false;
			isCharImmediatelyPostOperator = true;
		}
		if (isImmediatelyPostTemplate)
		{
			isImmediatelyPostTemplate = false;
			isCharImmediatelyPostTemplate = true;
		}
		if (isImmediatelyPostPointerOrReference)
		{
			isImmediatelyPostPointerOrReference = false;
			isCharImmediatelyPostPointerOrReference = true;
		}

		// reset isImmediatelyPostHeader information
		if (isImmediatelyPostHeader)
		{
			if (!handleImmediatelyPostHeaderSection())
				continue;
		}

		if (passedSemicolon)    // need to break the formattedLine
		{
			if (!handlePassedSemicolonSection())
				continue;
		}

		if (passedColon)
		{
			passedColon = false;
			if (parenStack->back() == 0
			        && !isBeforeAnyComment()
			        && (formattedLine.find_first_not_of(" \t") != std::string::npos))
			{
				shouldReparseCurrentChar = true;
				isInLineBreak = true;
				continue;
			}
		}

		// Check if in template declaration, e.g. foo<bar> or foo<bar,fig>
		if (!isInTemplate && currentChar == '<')
		{
			checkIfTemplateOpener();
		}

		// Check for break return type
		if ((size_t) charNum >= methodBreakCharNum && methodBreakLineNum == 0)
		{
			if ((size_t) charNum == methodBreakCharNum)
				isInLineBreak = true;
			methodBreakCharNum = std::string::npos;
			methodBreakLineNum = 0;
		}
		// Check for attach return type
		if ((size_t) charNum >= methodAttachCharNum && methodAttachLineNum == 0)
		{
			handleAttachedReturnTypes();
		}

		// handle parens
		if (currentChar == '(' || currentChar == '[' || (isInTemplate && currentChar == '<'))
		{
			handleOpenParens();
		}

		else if (currentChar == ')' || currentChar == ']' || (isInTemplate && currentChar == '>'))
		{
			handleClosedBracesOrParens();
		}

		// handle braces
		if (currentChar == '{' || currentChar == '}')
		{
			handleBraces();
			continue;
		}

		// #126
		if ( currentChar == '*' && shouldPadOperators &&
		        pointerAlignment != PTR_ALIGN_TYPE &&  // SF 557
		        peekNextChar() != '=' && // GH 81
		        ( currentHeader == &ASResource::AS_IF || currentHeader == &ASResource::AS_WHILE ||
		          currentHeader == &ASResource::AS_DO || currentHeader == &ASResource::AS_FOR)
		        && ( previousChar == ')' || std::isalpha(previousChar) )
		        && !isOperatorPaddingDisabled() )
		{
			appendSpacePad();
			appendOperator(ASResource::AS_MULT);
			goForward(0);
			appendSpaceAfter();
			continue;
		}

		if ((((previousCommandChar == '{' && isPreviousBraceBlockRelated)
		        || ((previousCommandChar == '}'
		             && !isImmediatelyPostEmptyBlock
		             && isPreviousBraceBlockRelated
		             && !isPreviousCharPostComment       // Fixes wrongly appended newlines after '}' immediately after comments
		             && peekNextChar() != ' '
		             && !isBraceType(previousBraceType, DEFINITION_TYPE))
		            && !isBraceType(braceTypeStack->back(), DEFINITION_TYPE)))
		        && isOkToBreakBlock(braceTypeStack->back()))
		        // check for array
		        || (previousCommandChar == '{'			// added 9/30/2010
		            && isBraceType(braceTypeStack->back(), ARRAY_TYPE)
		            && !isBraceType(braceTypeStack->back(), SINGLE_LINE_TYPE)
		            && isNonInStatementArray)
		        // check for pico one line braces
		        || (formattingStyle == STYLE_PICO
		            && (previousCommandChar == '{' && isPreviousBraceBlockRelated)
		            && isBraceType(braceTypeStack->back(), COMMAND_TYPE)
		            && isBraceType(braceTypeStack->back(), SINGLE_LINE_TYPE)
		            && braceFormatMode == RUN_IN_MODE)
		   )
		{
			handleBreakLine();
		}

		// reset block handling flags
		isImmediatelyPostEmptyBlock = false;

		// Objective-C method prefix with no return type
		if (isImmediatelyPostObjCMethodPrefix && currentChar != '(')
		{
			if (shouldPadMethodPrefix || shouldUnPadMethodPrefix)
				padObjCMethodPrefix();
			isImmediatelyPostObjCMethodPrefix = false;
		}

		// look for headers
		bool isPotentialHeader = isCharPotentialHeader(currentLine, charNum);

		if (isPotentialHeader && !isInTemplate && squareBracketCount == 0)
		{
			if (!handlePotentialHeader(newHeader))
				continue;
		}   // (isPotentialHeader && !isInTemplate)

		if (isInLineBreak)          // OK to break line here
		{
			breakLine();
			if (isInVirginLine)		// adjust for the first line
			{
				lineCommentNoBeautify = lineCommentNoIndent;
				lineCommentNoIndent = false;
				if (isImmediatelyPostPreprocessor)
				{
					isInIndentablePreproc = isIndentablePreprocessor;
					isIndentablePreprocessor = false;
				}
			}
		}

		if (previousNonWSChar == '}' || currentChar == ';')
		{
			handleEndOfBlock();
		}

		if (currentChar == ':'
		        && previousChar != ':'         // not part of '::'
		        && peekNextChar() != ':')      // not part of '::'
		{
			handleColonSection();
		}

		if (currentChar == '?')
			foundQuestionMark = true;

		if (isPotentialHeader && !isInTemplate)
		{
			handlePotentialHeaderPart2();
			continue;
		}   // (isPotentialHeader &&  !isInTemplate)

		// determine if this is an Objective-C statement

		if (currentChar == '@'
		        && isCStyle()
		        && (int) currentLine.length() > charNum + 1
		        && !std::isblank(currentLine[charNum + 1])
		        && isCharPotentialHeader(currentLine, charNum + 1)
		        && findKeyword(currentLine, charNum + 1, ASResource::AS_INTERFACE)
		        && isBraceType(braceTypeStack->back(), NULL_TYPE))
		{
			isInObjCInterface = true;
			std::string name = '@' + ASResource::AS_INTERFACE;
			appendSequence(name);
			goForward(name.length() - 1);
			continue;
		}
		if (currentChar == '@'
		        && isCStyle()
		        && (int) currentLine.length() > charNum + 1
		        && !std::isblank(currentLine[charNum + 1])
		        && isCharPotentialHeader(currentLine, charNum + 1)
		        && findKeyword(currentLine, charNum + 1, ASResource::AS_SELECTOR))
		{
			isInObjCSelector = true;
			std::string name = '@' + ASResource::AS_SELECTOR;
			appendSequence(name);
			goForward(name.length() - 1);
			continue;
		}
		if ((currentChar == '-' || currentChar == '+')
		        && isCStyle()
		        && (int) currentLine.find_first_not_of(" \t") == charNum
		        && !isInPotentialCalculation
		        && !isInObjCMethodDefinition
		        && (isBraceType(braceTypeStack->back(), NULL_TYPE)
		            || (isBraceType(braceTypeStack->back(), EXTERN_TYPE))))
		{
			isInObjCMethodDefinition = true;
			isImmediatelyPostObjCMethodPrefix = true;
			isInObjCParam = false;
			isInObjCInterface = false;
			if (getAlignMethodColon())
				objCColonAlign = findObjCColonAlignment();
			appendCurrentChar();
			continue;
		}

		// determine if this is a potential calculation

		bool isPotentialOperator = isCharPotentialOperator(currentChar);
		newHeader = nullptr;

		if (isPotentialOperator)
		{
			newHeader = findOperator(operators);

			handlePotentialOperator(newHeader);
		}

		// TODO check add flag to preserve space
		size_t lastNonWsChar = currentLine.find_last_not_of(" \t", charNum - 1);
		if (lastNonWsChar != std::string::npos && pointerAlignment == PTR_ALIGN_TYPE && !isGSCStyle() && !preserveWhitespace)
		{
			char lastChar = currentLine[lastNonWsChar];

			//if (lastChar != '(' && !isalpha(lastChar)) {
			//	formattedLine = rtrim(formattedLine);
			//}

			if (lastChar == ',')
			{
				formattedLine = rtrim(formattedLine);
				formattedLine += ' ';
			}

		}

		// process pointers and references
		// check newHeader to eliminate things like '&&' sequence
		if (newHeader != nullptr && !isJavaStyle()
		        && (newHeader == &ASResource::AS_MULT
		            || newHeader == &ASResource::AS_BIT_AND
		            || newHeader == &ASResource::AS_BIT_XOR
		            || newHeader == &ASResource::AS_AND)
		        && isPointerOrReference())
		{

			if (!isDereferenceOrAddressOf() && !isOperatorPaddingDisabled())
			{
				formatPointerOrReference();
			}
			else
			{
				appendOperator(*newHeader);
				goForward(newHeader->length() - 1);
			}
			isImmediatelyPostPointerOrReference = true;
			continue;
		}

		if ((shouldPadOperators || negationPadMode != NEGATION_PAD_NO_CHANGE) && newHeader != nullptr && !isOperatorPaddingDisabled())
		{
			padOperators(newHeader);
			continue;
		}

		// remove spaces before commas
		if (currentChar == ',')
		{
			const size_t len = formattedLine.length();
			size_t lastText = formattedLine.find_last_not_of(' ');
			if (lastText != std::string::npos && lastText < len - 1)
			{
				formattedLine.resize(lastText + 1);
				int size_diff = len - (lastText + 1);
				spacePadNum -= size_diff;
			}
		}

		// pad commas and semi-colons
		if (currentChar == ';'
		        || (currentChar == ',' && (shouldPadOperators || shouldPadCommas)))
		{
			char nextChar = ' ';
			if (charNum + 1 < (int) currentLine.length())
				nextChar = currentLine[charNum + 1];
			if (!std::isblank(nextChar)
			        && nextChar != '}'
			        && nextChar != ')'
			        && nextChar != ']'
			        && nextChar != '>'
			        && nextChar != ';'
			        && !isBeforeAnyComment()
			   )
			{
				appendCurrentChar();
				appendSpaceAfter();
				continue;
			}
		}

		// pad parens
		if (currentChar == '(' || currentChar == ')')
		{
			handleParens();
			continue;
		}

		//GL31
		bool isDoubleOpenBrackets = isGSCStyle() && currentChar == '[' && peekNextChar() == '[';

		if ((currentChar == '[' || currentChar == ']' ) && (shouldPadBracketsOutside || shouldPadBracketsInside || shouldUnPadBrackets) && !isDoubleOpenBrackets)
		{
			padParensOrBrackets('[', ']', false);
			continue;
		}

		// bypass the entire operator
		if (newHeader != nullptr)
		{
			appendOperator(*newHeader);
			goForward(newHeader->length() - 1);
			continue;
		}

		appendCurrentChar();

	}   // end of while loop  *  end of while loop  *  end of while loop  *  end of while loop

	// return a beautified (i.e. correctly indented) line.

	std::string beautifiedLine;
	size_t readyFormattedLineLength = trim(readyFormattedLine).length();
	bool isInNamespace = isBraceType(braceTypeStack->back(), NAMESPACE_TYPE);

	if (prependEmptyLine		// prepend a blank line before this formatted line
	        && readyFormattedLineLength > 0
	        && previousReadyFormattedLineLength > 0)
	{
		isLineReady = true;		// signal a waiting readyFormattedLine
		beautifiedLine = beautify("");
		previousReadyFormattedLineLength = 0;
		// call the enhancer for new empty lines
		enhancer->enhance(beautifiedLine, isInNamespace, isInPreprocessorBeautify, isInBeautifySQL);
	}
	else		// format the current formatted line
	{
		isLineReady = false;
		runInIndentContinuation = runInIndentChars;
		beautifiedLine = beautify(readyFormattedLine);
		previousReadyFormattedLineLength = readyFormattedLineLength;
		// the enhancer is not called for no-indent line comments
		if (!lineCommentNoBeautify && !isFormattingModeOff)
			enhancer->enhance(beautifiedLine, isInNamespace, isInPreprocessorBeautify, isInBeautifySQL);
		runInIndentChars = 0;
		lineCommentNoBeautify = lineCommentNoIndent;
		lineCommentNoIndent = false;
		isInIndentablePreproc = isIndentablePreprocessor;
		isIndentablePreprocessor = false;
		isElseHeaderIndent = elseHeaderFollowsComments;
		isCaseHeaderCommentIndent = caseHeaderFollowsComments;
		objCColonAlignSubsequent = objCColonAlign;
		if (isCharImmediatelyPostNonInStmt)
		{
			isNonInStatementArray = false;
			isCharImmediatelyPostNonInStmt = false;
		}
		isInPreprocessorBeautify = isInPreprocessor;	// used by ASEnhancer
		isInBeautifySQL = isInExecSQL;					// used by ASEnhancer
	}

	prependEmptyLine = false;
	assert(computeChecksumOut(beautifiedLine));
	return beautifiedLine;
}

/**
 * check if there are any indented lines ready to be read by nextLine()
 *
 * @return    are there any indented lines ready?
 */
bool ASFormatter::hasMoreLines() const
{
	return !endOfCodeReached;
}

/**
 * comparison function for BraceType enum
 */
bool ASFormatter::isBraceType(BraceType a, BraceType b) const
{
	if (a == NULL_TYPE || b == NULL_TYPE)
		return (a == b);
	return ((a & b) == b);
}

/**
 * set the formatting style.
 *
 * @param style         the formatting style.
 */
void ASFormatter::setFormattingStyle(FormatStyle style)
{
	formattingStyle = style;
}

/**
 * set the add braces mode.
 * options:

 *    true    braces added to headers for single line statements.
 *    false    braces NOT added to headers for single line statements.
 *
 * @param state         the add braces state.
 */
void ASFormatter::setAddBracesMode(bool state)
{
	shouldAddBraces = state;
}

/**
 * set the add one line braces mode.
 * options:
 *    true     one line braces added to headers for single line statements.
 *    false    one line braces NOT added to headers for single line statements.
 *
 * @param state         the add one line braces state.
 */
void ASFormatter::setAddOneLineBracesMode(bool state)
{
	shouldAddBraces = state ? 1 : 0;
	shouldAddOneLineBraces = state;
}

/**
 * set the remove braces mode.
 * options:
 *    true     braces removed from headers for single line statements.
 *    false    braces NOT removed from headers for single line statements.
 *
 * @param state         the remove braces state.
 */
void ASFormatter::setRemoveBracesMode(bool state)
{
	shouldRemoveBraces = state;
}

// retained for compatibility with release 2.06
// "Brackets" have been changed to "Braces" in 3.0
// it is referenced only by the old "bracket" options
void ASFormatter::setAddBracketsMode(bool state)
{
	setAddBracesMode(state);
}

// retained for compatibility with release 2.06
// "Brackets" have been changed to "Braces" in 3.0
// it is referenced only by the old "bracket" options
void ASFormatter::setAddOneLineBracketsMode(bool state)
{
	setAddOneLineBracesMode(state);
}

// retained for compatibility with release 2.06
// "Brackets" have been changed to "Braces" in 3.0
// it is referenced only by the old "bracket" options
void ASFormatter::setRemoveBracketsMode(bool state)
{
	setRemoveBracesMode(state);
}

// retained for compatibility with release 2.06
// "Brackets" have been changed to "Braces" in 3.0
// it is referenced only by the old "bracket" options
void ASFormatter::setBreakClosingHeaderBracketsMode(bool state)
{
	setBreakClosingHeaderBracesMode(state);
}

/**
 * set the brace formatting mode.
 * options:
 *
 * @param mode         the brace formatting mode.
 */
void ASFormatter::setBraceFormatMode(BraceMode mode)
{
	braceFormatMode = mode;
}

/**
 * set 'break after' mode for maximum code length
 *
 * @param state         the 'break after' mode.
 */
void ASFormatter::setBreakAfterMode(bool state)
{
	shouldBreakLineAfterLogical = state;
}

/**
 * set closing header brace breaking mode
 * options:
 *    true     braces just before closing headers (e.g. 'else', 'catch')
 *             will be broken, even if standard braces are attached.
 *    false    closing header braces will be treated as standard braces.
 *
 * @param state         the closing header brace breaking mode.
 */
void ASFormatter::setBreakClosingHeaderBracesMode(bool state)
{
	shouldBreakClosingHeaderBraces = state;
}

/**
 * set 'else if()' breaking mode
 * options:
 *    true     'else' headers will be broken from their succeeding 'if' headers.
 *    false    'else' headers will be attached to their succeeding 'if' headers.
 *
 * @param state         the 'else if()' breaking mode.
 */
void ASFormatter::setBreakElseIfsMode(bool state)
{
	shouldBreakElseIfs = state;
}

/**
* set comma padding mode.
* options:
*    true     statement commas and semicolons will be padded with spaces around them.
*    false    statement commas and semicolons will not be padded.
*
* @param state         the padding mode.
*/
void ASFormatter::setCommaPaddingMode(bool state)
{
	shouldPadCommas = state;
}

/**
 * set maximum code length
 *
 * @param max         the maximum code length.
 */
void ASFormatter::setMaxCodeLength(int max)
{
	maxCodeLength = max;
}

/**
 * set operator padding mode.
 * options:
 *    true     statement operators will be padded with spaces around them.
 *    false    statement operators will not be padded.
 *
 * @param state         the padding mode.
 */
void ASFormatter::setOperatorPaddingMode(bool state)
{
	shouldPadOperators = state;
}

/**
 * set negation padding mode.
 * @param state         the padding mode.
 */
void ASFormatter::setNegationPaddingMode(NegationPaddingMode mode)
{
	negationPadMode = mode;
}

/**
 * set include directive padding mode.
 * @param state         the padding mode.
 */
void ASFormatter::setIncludeDirectivePaddingMode(IncludeDirectivePaddingMode mode)
{
	includeDirectivePaddingMode = mode;
}

/**
 * set parenthesis outside padding mode.
 * options:
 *    true     statement parentheses will be padded with spaces around them.
 *    false    statement parentheses will not be padded.
 *
 * @param state         the padding mode.
 */
void ASFormatter::setParensOutsidePaddingMode(bool state)
{
	shouldPadParensOutside = state;
}

/**
 * set parenthesis inside padding mode.
 * options:
 *    true     statement parenthesis will be padded with spaces around them.
 *    false    statement parenthesis will not be padded.
 *
 * @param state         the padding mode.
 */
void ASFormatter::setParensInsidePaddingMode(bool state)
{
	shouldPadParensInside = state;
}

/**
 * set square brackets outside padding mode.
 * options:
 *    true     square brackets will be padded with spaces around them.
 *    false    square brackets will not be padded.
 *
 * @param state         the padding mode.
 */
void ASFormatter::setBracketsOutsidePaddingMode(bool state)
{
	shouldPadBracketsOutside = state;
}

/**
 * set square brackets inside padding mode.
 * options:
 *    true     square brackets will be padded with spaces around them.
 *    false    square brackets will not be padded.
 *
 * @param state         the padding mode.
 */
void ASFormatter::setBracketsInsidePaddingMode(bool state)
{
	shouldPadBracketsInside = state;
}

/**
 * set padding mode before one or more open parentheses.
 * options:
 *    true     first open parenthesis will be padded with a space before.
 *    false    first open parenthesis will not be padded.
 *
 * @param state         the padding mode.
 */
void ASFormatter::setParensFirstPaddingMode(bool state)
{
	shouldPadFirstParen = state;
}

/**
 * set padding mode for empty parentheses.
 * options:
 *    true     padding will be applied
 *    false    no padding (default)
 *
 * @param state         the padding mode.
 */
void ASFormatter::setEmptyParensPaddingMode(bool state)
{
	shouldPadEmptyParens = state;
}

/**
 * set header padding mode.
 * options:
 *    true     headers will be padded with spaces around them.
 *    false    headers will not be padded.
 *
 * @param state         the padding mode.
 */
void ASFormatter::setParensHeaderPaddingMode(bool state)
{
	shouldPadHeader = state;
}

/**
 * set parenthesis unpadding mode.
 * options:
 *    true     statement parenthesis will be unpadded with spaces removed around them.
 *    false    statement parenthesis will not be unpadded.
 *
 * @param state         the padding mode.
 */
void ASFormatter::setParensUnPaddingMode(bool state)
{
	shouldUnPadParens = state;
}

/**
 * set square brackets unpadding mode.
 * options:
 *    true     square brackets will be unpadded with spaces removed around them.
 *    false    square brackets will not be unpadded.
 *
 * @param state         the padding mode.
 */
void ASFormatter::setBracketsUnPaddingMode(bool state)
{
	shouldUnPadBrackets = state;
}

/**
* set the state of the preprocessor indentation option.
* If true, #ifdef blocks at level 0 will be indented.
*
* @param   state             state of option.
*/
void ASFormatter::setPreprocBlockIndent(bool state)
{
	shouldIndentPreprocBlock = state;
}

/**
 * Set strip comment prefix mode.
 * options:
 *    true     strip leading '*' in a comment.
 *    false    leading '*' in a comment will be left unchanged.
 *
 * @param state         the strip comment prefix mode.
 */
void ASFormatter::setStripCommentPrefix(bool state)
{
	shouldStripCommentPrefix = state;
}

/**
 * set objective-c '-' or '+' class prefix padding mode.
 * options:
 *    true     class prefix will be padded a spaces after them.
 *    false    class prefix will be left unchanged.
 *
 * @param state         the padding mode.
 */
void ASFormatter::setMethodPrefixPaddingMode(bool state)
{
	shouldPadMethodPrefix = state;
}

/**
 * set objective-c '-' or '+' class prefix unpadding mode.
 * options:
 *    true     class prefix will be unpadded with spaces after them removed.
 *    false    class prefix will left unchanged.
 *
 * @param state         the unpadding mode.
 */
void ASFormatter::setMethodPrefixUnPaddingMode(bool state)
{
	shouldUnPadMethodPrefix = state;
}

// set objective-c '-' or '+' return type padding mode.
void ASFormatter::setReturnTypePaddingMode(bool state)
{
	shouldPadReturnType = state;
}

// set objective-c '-' or '+' return type unpadding mode.
void ASFormatter::setReturnTypeUnPaddingMode(bool state)
{
	shouldUnPadReturnType = state;
}

// set objective-c method parameter type padding mode.
void ASFormatter::setParamTypePaddingMode(bool state)
{
	shouldPadParamType = state;
}

// set objective-c method parameter type unpadding mode.
void ASFormatter::setParamTypeUnPaddingMode(bool state)
{
	shouldUnPadParamType = state;
}

/**
 * set objective-c method colon padding mode.
 *
 * @param mode         objective-c colon padding mode.
 */
void ASFormatter::setObjCColonPaddingMode(ObjCColonPad mode)
{
	shouldPadMethodColon = true;
	objCColonPadMode = mode;
}

/**
 * set option to attach closing braces
 *
 * @param state        true = attach, false = don't attach.
 */
void ASFormatter::setAttachClosingBraceMode(bool state)
{
	attachClosingBraceMode = state;
}

/**
 * set option to attach class braces
 *
 * @param state        true = attach, false = use style default.
 */
void ASFormatter::setAttachClass(bool state)
{
	shouldAttachClass = state;
}

/**
 * set option to attach extern "C" braces
 *
 * @param state        true = attach, false = use style default.
 */
void ASFormatter::setAttachExternC(bool state)
{
	shouldAttachExternC = state;
}

/**
 * set option to attach namespace braces
 *
 * @param state        true = attach, false = use style default.
 */
void ASFormatter::setAttachNamespace(bool state)
{
	shouldAttachNamespace = state;
}

/**
 * set option to attach inline braces
 *
 * @param state        true = attach, false = use style default.
 */
void ASFormatter::setAttachInline(bool state)
{
	shouldAttachInline = state;
}

void ASFormatter::setAttachClosingWhile(bool state)
{
	shouldAttachClosingWhile = state;
}

/**
 * set option to break/not break one-line blocks
 *
 * @param state        true = break, false = don't break.
 */
void ASFormatter::setBreakOneLineBlocksMode(bool state)
{
	shouldBreakOneLineBlocks = state;
}

/**
* set one line headers breaking mode
*/
void ASFormatter::setBreakOneLineHeadersMode(bool state)
{
	shouldBreakOneLineHeaders = state;
}

/**
* set option to break/not break lines consisting of multiple statements.
*
* @param state        true = break, false = don't break.
*/
void ASFormatter::setBreakOneLineStatementsMode(bool state)
{
	shouldBreakOneLineStatements = state;
}

void ASFormatter::setCloseTemplatesMode(bool state)
{
	shouldCloseTemplates = state;
}

/**
 * set option to convert tabs to spaces.
 *
 * @param state        true = convert, false = don't convert.
 */
void ASFormatter::setTabSpaceConversionMode(bool state)
{
	shouldConvertTabs = state;
}

/**
 * set option to indent comments in column 1.
 *
 * @param state        true = indent, false = don't indent.
 */
void ASFormatter::setIndentCol1CommentsMode(bool state)
{
	shouldIndentCol1Comments = state;
}

/**
 * set option to force all line ends to a particular style.
 *
 * @param fmt           format enum value
 */
void ASFormatter::setLineEndFormat(LineEndFormat fmt)
{
	lineEnd = fmt;
}

/**
 * set option to break unrelated blocks of code with empty lines.
 *
 * @param state        true = convert, false = don't convert.
 */
void ASFormatter::setBreakBlocksMode(bool state)
{
	shouldBreakBlocks = state;
}

/**
 * set option to break closing header blocks of code (such as 'else', 'catch', ...) with empty lines.
 *
 * @param state        true = convert, false = don't convert.
 */
void ASFormatter::setBreakClosingHeaderBlocksMode(bool state)
{
	shouldBreakClosingHeaderBlocks = state;
}

/**
 * set option to delete empty lines.
 *
 * @param state        true = delete, false = don't delete.
 */
void ASFormatter::setDeleteEmptyLinesMode(bool state)
{
	shouldDeleteEmptyLines = state;
}

void ASFormatter::setBreakReturnType(bool state)
{
	shouldBreakReturnType = state;
}

void ASFormatter::setBreakReturnTypeDecl(bool state)
{
	shouldBreakReturnTypeDecl = state;
}

void ASFormatter::setAttachReturnType(bool state)
{
	shouldAttachReturnType = state;
}

void ASFormatter::setAttachReturnTypeDecl(bool state)
{
	shouldAttachReturnTypeDecl = state;
}

void ASFormatter::setSqueezeEmptyLinesNumber(int num)
{
	squeezeEmptyLineNum = num;
}


/**
 * set the pointer alignment.
 *
 * @param alignment    the pointer alignment.
 */
void ASFormatter::setPointerAlignment(PointerAlign alignment)
{
	pointerAlignment = alignment;
}

void ASFormatter::setReferenceAlignment(ReferenceAlign alignment)
{
	referenceAlignment = alignment;
}

/**
 * jump over several characters.
 *
 * @param i       the number of characters to jump over.
 */
void ASFormatter::goForward(int i)
{
	while (--i >= 0)
		getNextChar();
}

/**
 * peek at the next unread character.
 *
 * @return     the next unread character.
 */
char ASFormatter::peekNextChar() const
{
	char ch = ' ';
	size_t peekNum = currentLine.find_first_not_of(" \t", charNum + 1);

	if (peekNum == std::string::npos)
		return ch;

	ch = currentLine[peekNum];

	return ch;
}

/**
 * check if current placement is before a comment
 *
 * @return     is before a comment.
 */
bool ASFormatter::isBeforeComment() const
{
	bool foundComment = false;
	size_t peekNum = currentLine.find_first_not_of(" \t", charNum + 1);

	if (peekNum == std::string::npos)
		return foundComment;

	foundComment = (currentLine.compare(peekNum, 2, "/*") == 0);

	return foundComment;
}

/**
 * check if current placement is before a comment or line-comment
 *
 * @return     is before a comment or line-comment.
 */
bool ASFormatter::isBeforeAnyComment() const
{
	bool foundComment = false;
	size_t peekNum = currentLine.find_first_not_of(" \t", charNum + 1);

	if (peekNum == std::string::npos)
		return foundComment;

	foundComment = (currentLine.compare(peekNum, 2, "/*") == 0
	                || currentLine.compare(peekNum, 2, "//") == 0);

	return foundComment;
}

/**
 * check if current placement is before a comment or line-comment
 * if a block comment it must be at the end of the line
 *
 * @return     is before a comment or line-comment.
 */
bool ASFormatter::isBeforeAnyLineEndComment(int startPos) const
{
	bool foundLineEndComment = false;
	size_t peekNum = currentLine.find_first_not_of(" \t", startPos + 1);

	if (peekNum != std::string::npos)
	{
		if (currentLine.compare(peekNum, 2, "//") == 0)
			foundLineEndComment = true;
		else if (currentLine.compare(peekNum, 2, "/*") == 0)
		{
			// comment must be closed on this line with nothing after it
			size_t endNum = currentLine.find("*/", peekNum + 2);
			if (endNum != std::string::npos)
			{
				size_t nextChar = currentLine.find_first_not_of(" \t", endNum + 2);
				if (nextChar == std::string::npos)
					foundLineEndComment = true;
			}
		}
	}
	return foundLineEndComment;
}

/**
 * check if current placement is before a comment followed by a line-comment
 *
 * @return     is before a multiple line-end comment.
 */
bool ASFormatter::isBeforeMultipleLineEndComments(int startPos) const
{
	bool foundMultipleLineEndComment = false;
	size_t peekNum = currentLine.find_first_not_of(" \t", startPos + 1);

	if (peekNum != std::string::npos)
	{
		if (currentLine.compare(peekNum, 2, "/*") == 0)
		{
			// comment must be closed on this line with nothing after it
			size_t endNum = currentLine.find("*/", peekNum + 2);
			if (endNum != std::string::npos)
			{
				size_t nextChar = currentLine.find_first_not_of(" \t", endNum + 2);
				if (nextChar != std::string::npos
				        && currentLine.compare(nextChar, 2, "//") == 0)
					foundMultipleLineEndComment = true;
			}
		}
	}
	return foundMultipleLineEndComment;
}

/**
 * get the next character, increasing the current placement in the process.
 * the new character is inserted into the variable currentChar.
 *
 * @return   whether succeeded to receive the new character.
 */
bool ASFormatter::getNextChar()
{
	isInLineBreak = false;
	previousChar = currentChar;

	if (!std::isblank(currentChar))
	{
		previousNonWSChar = currentChar;
		if (!isInComment && !isInLineComment && !isInQuote
		        && !isImmediatelyPostComment
		        && !isImmediatelyPostLineComment
		        && !isInPreprocessor
		        && !isSequenceReached(ASResource::AS_OPEN_COMMENT)
		        && !(isGSCStyle() && isSequenceReached(ASResource::AS_GSC_OPEN_COMMENT))
		        && !isSequenceReached(ASResource::AS_OPEN_LINE_COMMENT))
			previousCommandChar = currentChar;
	}

	if (charNum + 1 < (int) currentLine.length()
	        && (!std::isblank(peekNextChar()) || isInComment || isInLineComment))
	{
		currentChar = currentLine[++charNum];
		if (currentChar == '\t' && shouldConvertTabs)
			convertTabToSpaces();

		return true;
	}

	// end of line has been reached
	return getNextLine();
}

/**
 * get the next line of input, increasing the current placement in the process.
 *
 * @param emptyLineWasDeleted         an empty line was deleted.
 * @return   whether succeeded in reading the next line.
 */
bool ASFormatter::getNextLine(bool emptyLineWasDeleted /*false*/)
{
	if (!sourceIterator->hasMoreLines())
	{
		endOfCodeReached = true;
		return false;
	}
	if (appendOpeningBrace)
		currentLine = "{";		// append brace that was removed from the previous line
	else
	{
		currentLine = sourceIterator->nextLine(emptyLineWasDeleted);
		assert(computeChecksumIn(currentLine));
	}

	// reset variables for new line
	inLineNumber++;
	if (endOfAsmReached)
		endOfAsmReached = isInAsmBlock = isInAsm = false;
	shouldKeepLineUnbroken = false;
	isInCommentStartLine = false;
	isInCase = false;
	isInAsmOneLine = false;
	isHeaderInMultiStatementLine = false;
	isInQuoteContinuation = isInVerbatimQuote || haveLineContinuationChar;
	haveLineContinuationChar = false;
	isImmediatelyPostEmptyLine = lineIsEmpty;
	previousChar = ' ';

	if (currentLine.empty())
	{
		//#574 avoid deletion of empty lines after continuation
		if (!isInComment && previousNonWSChar == '\\')
		{
			isInPreprocessor = true;
			return false;
		}

		isInContinuedPreProc = false;
		currentLine = std::string(" ");        // a null is inserted if this is not done
	}

	if (methodBreakLineNum > 0)
		--methodBreakLineNum;
	if (methodAttachLineNum > 0)
		--methodAttachLineNum;

	// unless reading in the first line of the file, break a new line.
	if (!isVirgin)
		isInLineBreak = true;
	else
		isVirgin = false;

	if (isImmediatelyPostNonInStmt)
	{
		isCharImmediatelyPostNonInStmt = true;
		isImmediatelyPostNonInStmt = false;
	}

	// check if is in preprocessor before line trimming
	// a blank line after a \ will remove the flag
	isImmediatelyPostPreprocessor = isInPreprocessor;

	if (!isInComment
	        && (previousNonWSChar != '\\'
	            || isEmptyLine(currentLine)))
	{
		isInPreprocessor = false;
		isInPreprocessorDefineDef = false;
	}

	if (passedSemicolon)
		isInExecSQL = false;
	initNewLine();

	currentChar = currentLine[charNum];
	if (isInBraceRunIn && previousNonWSChar == '{' && !isInComment)
		isInLineBreak = false;
	isInBraceRunIn = false;

	if (currentChar == '\t' && shouldConvertTabs)
		convertTabToSpaces();

	// check for an empty line inside a command brace.
	// if yes then read the next line (calls getNextLine recursively).
	// must be after initNewLine.
	if (shouldDeleteEmptyLines
	        && lineIsEmpty
	        && isBraceType((*braceTypeStack)[braceTypeStack->size() - 1], COMMAND_TYPE))
	{
		if (!shouldBreakBlocks || previousNonWSChar == '{' || !commentAndHeaderFollows())
		{
			isInPreprocessor = isImmediatelyPostPreprocessor;		// restore
			lineIsEmpty = false;
			return getNextLine(true);
		}
	}

	if ( ++squeezeEmptyLineCount > squeezeEmptyLineNum && lineIsEmpty && isImmediatelyPostEmptyLine)
	{
		isInPreprocessor = isImmediatelyPostPreprocessor;		// restore
		return getNextLine(true);
	}

	return true;
}

/**
 * jump over the leading white space in the current line,
 * IF the line does not begin a comment or is in a preprocessor definition.
 */
void ASFormatter::initNewLine()
{
	size_t len = currentLine.length();
	size_t tabSize = getTabLength();
	charNum = 0;

	// don't trim these
	if (isInQuoteContinuation
	        || (isInPreprocessor && !getPreprocDefineIndent()))
		return;

	// SQL continuation lines must be adjusted so the leading spaces
	// is equivalent to the opening EXEC SQL
	if (isInExecSQL)
	{
		// replace leading tabs with spaces
		// so that continuation indent will be spaces
		size_t tabCount_ = 0;
		size_t i;
		for (i = 0; i < currentLine.length(); i++)
		{
			if (!std::isblank(currentLine[i]))		// stop at first text
				break;
			if (currentLine[i] == '\t')
			{
				size_t numSpaces = tabSize - ((tabCount_ + i) % tabSize);
				currentLine.replace(i, 1, numSpaces, ' ');
				tabCount_++;
				i += tabSize - 1;
			}
		}
		// this will correct the format if EXEC SQL is not a hanging indent
		trimContinuationLine();
		return;
	}

	// comment continuation lines must be adjusted so the leading spaces
	// is equivalent to the opening comment
	if (isInComment)
	{
		if (noTrimCommentContinuation)
			leadingSpaces = tabIncrementIn = 0;
		trimContinuationLine();
		return;
	}

	// compute leading spaces
	isImmediatelyPostCommentOnly = lineIsLineCommentOnly || lineEndsInCommentOnly;
	lineIsCommentOnly = false;
	lineIsLineCommentOnly = false;
	lineEndsInCommentOnly = false;
	doesLineStartComment = false;
	currentLineBeginsWithBrace = false;
	lineIsEmpty = false;
	currentLineFirstBraceNum = std::string::npos;
	tabIncrementIn = 0;

	// bypass whitespace at the start of a line
	// preprocessor tabs are replaced later in the program
	for (charNum = 0; std::isblank(currentLine[charNum]) && charNum + 1 < (int) len; charNum++)
	{
		if (currentLine[charNum] == '\t'
		        && (!isInPreprocessor || isInPreprocessorDefineDef))
			tabIncrementIn += tabSize - 1 - ((tabIncrementIn + charNum) % tabSize);
	}
	leadingSpaces = charNum + tabIncrementIn;

	if (isSequenceReached(ASResource::AS_OPEN_COMMENT) || (isGSCStyle() && isSequenceReached(ASResource::AS_GSC_OPEN_COMMENT)))
	{
		doesLineStartComment = true;
		if ((int) currentLine.length() > charNum + 2
		        && currentLine.find("*/", charNum + 2) != std::string::npos)
			lineIsCommentOnly = true;
	}
	else if (isSequenceReached(ASResource::AS_OPEN_LINE_COMMENT))
	{
		lineIsLineCommentOnly = true;
	}
	else if (isSequenceReached("{"))
	{
		currentLineBeginsWithBrace = true;
		currentLineFirstBraceNum = charNum;
		size_t firstText = currentLine.find_first_not_of(" \t", charNum + 1);
		if (firstText != std::string::npos)
		{
			if (currentLine.compare(firstText, 2, "//") == 0)
				lineIsLineCommentOnly = true;
			else if (currentLine.compare(firstText, 2, "/*") == 0
			         || isExecSQL(currentLine, firstText))
			{
				// get the extra adjustment
				size_t j;
				for (j = charNum + 1; j < firstText && std::isblank(currentLine[j]); j++)
				{
					if (currentLine[j] == '\t')
						tabIncrementIn += tabSize - 1 - ((tabIncrementIn + j) % tabSize);
				}
				leadingSpaces = j + tabIncrementIn;
				if (currentLine.compare(firstText, 2, "/*") == 0)
					doesLineStartComment = true;
			}
		}
	}
	else if (std::isblank(currentLine[charNum]) && !(charNum + 1 < (int) currentLine.length()))
	{
		lineIsEmpty = true;
		if (!isImmediatelyPostEmptyLine  )
		{
			squeezeEmptyLineCount = 0;
		}
	}

	// do not trim indented preprocessor define (except for comment continuation lines)
	if (isInPreprocessor)
	{
		if (!doesLineStartComment)
			leadingSpaces = 0;
		charNum = 0;
	}
}

/**
 * Append a character to the current formatted line.
 * The formattedLine split points are updated.
 *
 * @param ch               the character to append.
 * @param canBreakLine     if true, a registered line-break
 */
void ASFormatter::appendChar(char ch, bool canBreakLine)
{
	if (canBreakLine && isInLineBreak)
		breakLine();

	formattedLine.append(1, ch);
	isImmediatelyPostCommentOnly = false;
	if (maxCodeLength != std::string::npos)
	{
		// These compares reduce the frequency of function calls.
		if (isOkToSplitFormattedLine())
			updateFormattedLineSplitPoints(ch);
		if (formattedLine.length() > maxCodeLength)
			testForTimeToSplitFormattedLine();
	}
}

/**
 * Append a std::string sequence to the current formatted line.
 * The formattedLine split points are NOT updated.
 * But the formattedLine is checked for time to split.
 *
 * @param sequence         the sequence to append.
 * @param canBreakLine     if true, a registered line-break
 */
void ASFormatter::appendSequence(std::string_view sequence, bool canBreakLine)
{
	if (canBreakLine && isInLineBreak)
		breakLine();
	formattedLine.append(sequence);
	if (formattedLine.length() > maxCodeLength)
		testForTimeToSplitFormattedLine();
}

/**
 * Append an operator sequence to the current formatted line.
 * The formattedLine split points are updated.
 *
 * @param sequence         the sequence to append.
 * @param canBreakLine     if true, a registered line-break
 */
void ASFormatter::appendOperator(std::string_view sequence, bool canBreakLine)
{
	if (canBreakLine && isInLineBreak)
		breakLine();
	formattedLine.append(sequence);
	if (maxCodeLength != std::string::npos)
	{
		// These compares reduce the frequency of function calls.
		if (isOkToSplitFormattedLine())
			updateFormattedLineSplitPointsOperator(sequence);
		if (formattedLine.length() > maxCodeLength)
			testForTimeToSplitFormattedLine();
	}
}

/**
 * append a space to the current formattedline, UNLESS the
 * last character is already a white-space character.
 */
void ASFormatter::appendSpacePad()
{
	int len = formattedLine.length();
	if (len > 0 && !std::isblank(formattedLine[len - 1]))
	{
		formattedLine.append(1, ' ');
		spacePadNum++;
		if (maxCodeLength != std::string::npos)
		{
			// These compares reduce the frequency of function calls.
			if (isOkToSplitFormattedLine())
				updateFormattedLineSplitPoints(' ');
			if (formattedLine.length() > maxCodeLength)
				testForTimeToSplitFormattedLine();
		}
	}
}

/**
 * append a space to the current formattedline, UNLESS the
 * next character is already a white-space character.
 */
void ASFormatter::appendSpaceAfter()
{
	int len = currentLine.length();
	if (charNum + 1 < len && !std::isblank(currentLine[charNum + 1]))
	{
		formattedLine.append(1, ' ');
		spacePadNum++;
		if (maxCodeLength != std::string::npos)
		{
			// These compares reduce the frequency of function calls.
			if (isOkToSplitFormattedLine())
				updateFormattedLineSplitPoints(' ');
			if (formattedLine.length() > maxCodeLength)
				testForTimeToSplitFormattedLine();
		}
	}
}

/**
 * register a line break for the formatted line.
 */
void ASFormatter::breakLine(bool isSplitLine /*false*/)
{
	isLineReady = true;
	isInLineBreak = false;
	spacePadNum = nextLineSpacePadNum;
	nextLineSpacePadNum = 0;
	readyFormattedLine = formattedLine;
	formattedLine.erase();
	// queue an empty line prepend request if one exists
	prependEmptyLine = isPrependPostBlockEmptyLineRequested;

	if (!isSplitLine)
	{
		formattedLineCommentNum = std::string::npos;
		clearFormattedLineSplitPoints();

		if (isAppendPostBlockEmptyLineRequested)
		{
			isAppendPostBlockEmptyLineRequested = false;
			isPrependPostBlockEmptyLineRequested = true;
		}
		else
			isPrependPostBlockEmptyLineRequested = false;
	}
}

/**
 * check if the currently reached open-brace (i.e. '{')
 * opens a:
 * - a definition type block (such as a class or namespace),
 * - a command block (such as a method block)
 * - a static array
 * this method takes for granted that the current character
 * is an opening brace.
 *
 * @return    the type of the opened block.
 */
BraceType ASFormatter::getBraceType()
{
	assert(currentChar == '{');

	BraceType returnVal = NULL_TYPE;

	if ((previousNonWSChar == '='
	        || isBraceType(braceTypeStack->back(), ARRAY_TYPE))
	        && previousCommandChar != ')'
	        && !isNonParenHeader)
		returnVal = ARRAY_TYPE;
	else if (foundPreDefinitionHeader && previousCommandChar != ')')
	{
		returnVal = DEFINITION_TYPE;
		if (foundNamespaceHeader)
			returnVal = (BraceType)(returnVal | NAMESPACE_TYPE);
		else if (foundClassHeader)
			returnVal = (BraceType)(returnVal | CLASS_TYPE);
		else if (foundStructHeader)
			returnVal = (BraceType)(returnVal | STRUCT_TYPE);
		else if (foundInterfaceHeader)
			returnVal = (BraceType)(returnVal | INTERFACE_TYPE);
	}
	else if (isInEnum)
	{
		returnVal = (BraceType)(ARRAY_TYPE | ENUM_TYPE);
	}
	else if (isSharpStyle() &&
	         !isOneLineBlockReached(currentLine, charNum) &&
	         (currentHeader == &ASResource::AS_IF || currentHeader == &ASResource::AS_WHILE
	          || currentHeader == &ASResource::AS_USING || currentHeader == &ASResource::AS_WHILE
	          || currentHeader == &ASResource::AS_FOR  || currentHeader == &ASResource::AS_FOREACH) )   // GH16
	{
		returnVal = (BraceType) COMMAND_TYPE;
	}
	else
	{
		bool isCommandType = (foundPreCommandHeader
		                      || foundPreCommandMacro
		                      || (currentHeader != nullptr && isNonParenHeader)
		                      || (previousCommandChar == ')' && !isInAllocator)
		                      || (previousCommandChar == ':' && !foundQuestionMark)
		                      || (previousCommandChar == ';')
		                      || ((previousCommandChar == '{' || previousCommandChar == '}')
		                          && isPreviousBraceBlockRelated)
		                      || (isInClassInitializer
		                          && ((!isLegalNameChar(previousNonWSChar) && previousNonWSChar != '(')
		                              || foundPreCommandHeader))
		                      || foundTrailingReturnType
		                      || isInObjCMethodDefinition
		                      || isInObjCInterface
		                      || isJavaStaticConstructor
		                      || isSharpDelegate);
		// C# methods containing 'get', 'set', 'add', and 'remove' do NOT end with parens
		if (!isCommandType && isSharpStyle() && isNextWordSharpNonParenHeader(charNum + 1))
		{
			isCommandType = true;
			isSharpAccessor = true;
		}

		if (isInExternC)
			returnVal = (isCommandType ? COMMAND_TYPE : EXTERN_TYPE);
		else
			returnVal = (isCommandType ? COMMAND_TYPE : ARRAY_TYPE);
	}

	int foundOneLineBlock = isOneLineBlockReached(currentLine, charNum);

	if (foundOneLineBlock == 2 && returnVal == COMMAND_TYPE)
		returnVal = ARRAY_TYPE;

	if (foundOneLineBlock > 0)
	{
		returnVal = (BraceType) (returnVal | SINGLE_LINE_TYPE);
		if (breakCurrentOneLineBlock)
			returnVal = (BraceType) (returnVal | BREAK_BLOCK_TYPE);
		if (foundOneLineBlock == 3)
			returnVal = (BraceType)(returnVal | EMPTY_BLOCK_TYPE);
	}

	if (isBraceType(returnVal, ARRAY_TYPE))
	{
		if (isNonInStatementArrayBrace())
		{
			returnVal = (BraceType)(returnVal | ARRAY_NIS_TYPE);
			isNonInStatementArray = true;
			isImmediatelyPostNonInStmt = false;		// in case of "},{"
			nonInStatementBrace = formattedLine.length() - 1;
		}
		if (isUniformInitializerBrace())
			returnVal = (BraceType)(returnVal | INIT_TYPE);
	}


	return returnVal;
}

/**
* check if a colon is a class initializer separator
*
* @return        whether it is a class initializer separator
*/
bool ASFormatter::isClassInitializer() const
{
	assert(currentChar == ':');
	assert(previousChar != ':' && peekNextChar() != ':'); // not part of '::'

	// Early exit if conditions that prevent class initializer detection are met
	if (foundQuestionMark || parenStack->back() > 0 || isInEnum)
	{
		return false;
	}

	// Check if we are in a C-style class constructor initializer
	bool isCStyleInitializer = isCStyle() &&
	                           !isInCase &&
	                           (previousCommandChar == ')' || foundPreCommandHeader);

	return isCStyleInitializer;
}


/**
 * check if a line is empty
 *
 * @return        whether line is empty
 */
bool ASFormatter::isEmptyLine(std::string_view line) const
{
	return line.find_first_not_of(" \t") == std::string::npos;
}

/**
 * Check if the following text is "C" as in extern "C".
 *
 * @return        whether the statement is extern "C"
 */
bool ASFormatter::isExternC() const
{
	// charNum should be at 'extern'
	assert(!std::isblank(currentLine[charNum]));
	size_t startQuote = currentLine.find_first_of(" \t\"", charNum);
	if (startQuote == std::string::npos)
		return false;
	startQuote = currentLine.find_first_not_of(" \t", startQuote);
	if (startQuote == std::string::npos)
		return false;
	if (currentLine.compare(startQuote, 3, "\"C\"") != 0)
		return false;
	return true;
}

/**
 * Check if the currently reached '*', '&' or '^' character is
 * a pointer-or-reference symbol, or another operator.
 * A pointer dereference (*) or an "address of" character (&)
 * counts as a pointer or reference because it is not an
 * arithmetic operator.
 *
 * @return        whether current character is a reference-or-pointer
 */
bool ASFormatter::isPointerOrReference() const
{
	assert(currentChar == '*' || currentChar == '&' || currentChar == '^');

	if (isJavaStyle())
		return false;

	if (isCharImmediatelyPostOperator)
		return false;

	// get the last legal word (may be a number)
	std::string lastWord = getPreviousWord(currentLine, charNum);
	if (lastWord.empty())
		lastWord = " ";

	// check for preceding or following numeric values
	std::string nextText = peekNextText(currentLine.substr(charNum + 1));
	if (nextText.empty())
		nextText = " ";
	if (isDigit(lastWord[0])
	        || isDigit(nextText[0])
	        || nextText[0] == '!'
	        || nextText[0] == '~')
		return false;

	// check for multiply then a dereference (a * *b)
	char nextChar = peekNextChar();
	if (currentChar == '*'
	        && nextChar == '*'
	        && !isPointerToPointer(currentLine, charNum))
		return false;

	if ((foundCastOperator && nextChar == '>')
	        || isPointerOrReferenceVariable(lastWord))
	{
		return true;
	}

	if (pointerAlignment == PTR_ALIGN_TYPE
	        && !shouldPadOperators //TODO 578
	        && !isPointerOrReferenceVariable(lastWord))
	{
		return false;
	}

	if (isInClassInitializer
	        && previousNonWSChar != '('
	        && previousNonWSChar != '{'
	        && previousCommandChar != ','
	        && nextChar != ')'
	        && nextChar != '}')
		return false;

	//check for rvalue reference
	if (currentChar == '&' && nextChar == '&')
	{
		if (lastWord == ASResource::AS_AUTO)
			return true;
		if (previousNonWSChar == '>')
			return true;
		std::string followingText;
		if ((int) currentLine.length() > charNum + 2)
			followingText = peekNextText(currentLine.substr(charNum + 2));
		if (!followingText.empty() && followingText[0] == ')')
			return true;
		if (currentHeader != nullptr || isInPotentialCalculation)
			return false;
		if (parenStack->back() > 0 && isBraceType(braceTypeStack->back(), COMMAND_TYPE))
			return false;
		return true;
	}

	if (nextChar == '*'
	        || previousNonWSChar == '='
	        || previousNonWSChar == '('
	        || previousNonWSChar == '['
	        || isCharImmediatelyPostReturn
	        || isInTemplate
	        || isCharImmediatelyPostTemplate
	        || currentHeader == &ASResource::AS_CATCH
	        || currentHeader == &ASResource::AS_FOREACH
	        || currentHeader == &ASResource::AS_QFOREACH)
	{
		return true;
	}


	if (isBraceType(braceTypeStack->back(), ARRAY_TYPE)
	        && isLegalNameChar(lastWord[0])
	        && isLegalNameChar(nextChar)
	        && previousNonWSChar != ')')
	{
		if (isArrayOperator())
			return false;
	}

	// checks on operators in parens
	if (parenStack->back() > 0
	        && isLegalNameChar(lastWord[0])
	        && isLegalNameChar(nextChar))
	{
		// if followed by an assignment it is a pointer or reference
		// if followed by semicolon it is a pointer or reference in range-based for
		const std::string* followingOperator = getFollowingOperator();
		if (followingOperator != nullptr
		        && followingOperator != &ASResource::AS_MULT
		        && followingOperator != &ASResource::AS_BIT_AND)
		{
			if (followingOperator == &ASResource::AS_ASSIGN || followingOperator == &ASResource::AS_COLON)
			{
				return true;
			}

			return false;
		}

		if (isBraceType(braceTypeStack->back(), COMMAND_TYPE)
		        || squareBracketCount > 0)
			return false;
		return true;
	}

	// checks on operators in parens with following '('
	std::set<char> disallowedChars = {',', '(', '!', '&', '*', '|'};

	if (parenStack->back() > 0
	        && nextChar == '('
	        && disallowedChars.find(previousNonWSChar) == disallowedChars.end())
		return false;

	if (nextChar == '-'
	        || nextChar == '+')
	{
		size_t nextNum = currentLine.find_first_not_of(" \t", charNum + 1);
		if (nextNum != std::string::npos)
		{
			if (currentLine.compare(nextNum, 2, "++") != 0
			        && currentLine.compare(nextNum, 2, "--") != 0)
				return false;
		}
	}

	bool isPR = (!isInPotentialCalculation
	             || (!isLegalNameChar(previousNonWSChar)
	                 && !(previousNonWSChar == ')' && nextChar == '(')
	                 && !(previousNonWSChar == ')' && currentChar == '*' && !isImmediatelyPostCast())
	                 && previousNonWSChar != ']')
	             || (!std::isblank(nextChar)
	                 && nextChar != '-'
	                 && nextChar != '('
	                 && nextChar != '['
	                 && !isLegalNameChar(nextChar))
	            );

	return isPR;
}

/**
 * Check if the currently reached  '*' or '&' character is
 * a dereferenced pointer or "address of" symbol.
 * NOTE: this MUST be a pointer or reference as determined by
 * the function isPointerOrReference().
 *
 * @return        whether current character is a dereference or address of
 */
bool ASFormatter::isDereferenceOrAddressOf() const
{
	assert(currentChar == '*' || currentChar == '&' || currentChar == '^');

	if (isCharImmediatelyPostTemplate)
		return false;

	// https://sourceforge.net/p/astyle/bugs/537/
	// https://sourceforge.net/p/astyle/bugs/552/
	if ( previousNonWSChar == ',' && parenthesesCount <= 0 && currentChar != '&')
	{
		return false;
	}

	if ( currentChar == '*' && pointerAlignment == PTR_ALIGN_NAME )
	{
		size_t openParen = currentLine.rfind('(', charNum);
		if (openParen != std::string::npos)
		{
			return true;
		}
	}

	std::set<char> allowedChars = {'=', '.', '{', '>', '<', '?'};

	if ( allowedChars.find(previousNonWSChar) != allowedChars.end()
	        || (previousNonWSChar == ',' && currentChar == '&')  // #537, #552
	        || isCharImmediatelyPostLineComment
	        || isCharImmediatelyPostComment
	        || isCharImmediatelyPostReturn)
		return true;

	char nextChar = peekNextChar();
	if (currentChar == '*' && nextChar == '*')
	{
		if (previousNonWSChar == '(')
			return true;
		if ((int) currentLine.length() < charNum + 2)
			return true;
		return false;
	}

	if (currentChar == '&' && nextChar == '&')
	{
		if (previousNonWSChar == '(' || isInTemplate)
			return true;
		if ((int) currentLine.length() < charNum + 2)
			return true;
		return false;
	}

	if (previousNonWSChar == '(' && currentChar == '&' && pointerAlignment == PTR_ALIGN_TYPE)
	{
		return true;
	}

	// check first char on the line
	if (charNum == (int) currentLine.find_first_not_of(" \t")
	        && (isBraceType(braceTypeStack->back(), COMMAND_TYPE)
	            || parenStack->back() != 0))
		return true;

	std::string nextText = peekNextText(currentLine.substr(charNum + 1));
	if (!nextText.empty())
	{
		if (nextText[0] == ')' || nextText[0] == '>'
		        || nextText[0] == ',' || nextText[0] == '=')
			return false;
		if (nextText[0] == ';')
			return true;
	}
	// check for reference to a pointer *&
	if ((currentChar == '*' && nextChar == '&')
	        || (previousNonWSChar == '*' && currentChar == '&'))
		return false;

	if (!isBraceType(braceTypeStack->back(), COMMAND_TYPE)
	        && parenStack->back() == 0)
		return false;
	std::string lastWord = getPreviousWord(currentLine, charNum);
	if (lastWord == "else" || lastWord == "delete")
		return true;

	bool isDA = (!(isLegalNameChar(previousNonWSChar) || previousNonWSChar == '>')          // TODO GH14
	             || (!nextText.empty() && !isLegalNameChar(nextText[0]) && nextText[0] != '/')
	             || (ispunct((unsigned char)previousNonWSChar) && previousNonWSChar != '.') // TODO GH14
	             || isCharImmediatelyPostReturn
	             || !isPointerOrReferenceVariable(lastWord));

	return isDA;
}

/**
 * Check if the currently reached  '*' or '&' character is
 * centered with one space on each side.
 * Only spaces are checked, not tabs.
 * If true then a space will be deleted on the output.
 *
 * @return        whether current character is centered.
 */
bool ASFormatter::isPointerOrReferenceCentered() const
{
	assert(currentChar == '*' || currentChar == '&' || currentChar == '^');

	int prNum = charNum;
	int lineLength = (int) currentLine.length();

	// check for end of line
	if (peekNextChar() == ' ')
		return false;

	// check space before
	if (prNum < 1
	        || currentLine[prNum - 1] != ' ')
		return false;

	// check no space before that
	if (prNum < 2
	        || currentLine[prNum - 2] == ' ')
		return false;

	// check for ** or &&
	if (prNum + 1 < lineLength
	        && (currentLine[prNum + 1] == '*' || currentLine[prNum + 1] == '&'))
		prNum++;

	// check space after
	if (prNum + 1 <= lineLength
	        && currentLine[prNum + 1] != ' ')
		return false;

	// check no space after that
	if (prNum + 2 < lineLength
	        && currentLine[prNum + 2] == ' ')
		return false;

	return true;
}

/**
 * Check if a word is a pointer or reference variable type.
 *
 * @return        whether word is a pointer or reference variable.
 */
bool ASFormatter::isPointerOrReferenceVariable(std::string_view word) const
{
	assert(currentChar == '*' || currentChar == '&' || currentChar == '^');

	// to avoid problem with multiplications - we need LSP

	bool retval = false;
	if (word == "char"
	        || word == "std::string"
	        || word == "String"
	        || word == "NSString"
	        || word == "int"
	        || word == "void"
	        || word == "short"
	        || word == "long"
	        || word == "double"
	        || word == "float"
	        || (word.length() >= 6     // check end of word for _t
	            && word.compare(word.length() - 2, 2, "_t") == 0)
	   )
		retval = true;

	// check for C# object type "x is std::string"
	if (retval && isSharpStyle())
	{
		// find the word previous to the 'word' parameter
		std::string prevWord;
		size_t wordStart = currentLine.rfind(word, charNum);
		if (wordStart != std::string::npos)
			prevWord = getPreviousWord(currentLine, wordStart);
		if (prevWord == "is")
			retval = false;
	}

	return retval;
}

/**
 * Check if * * is a pointer to a pointer or a multiply then a dereference.
 *
 * @return        true if a pointer *.
 */
bool ASFormatter::isPointerToPointer(std::string_view line, int currPos) const
{
	assert(line[currPos] == '*' && peekNextChar() == '*');
	if ((int) line.length() > currPos + 1 && line[currPos + 1] == '*')
		return true;
	size_t nextText = line.find_first_not_of(" \t", currPos + 1);
	if (nextText == std::string::npos || line[nextText] != '*')
		return false;
	size_t nextText2 = line.find_first_not_of(" \t", nextText + 1);
	if (nextText == std::string::npos)
		return false;
	if (line[nextText2] == ')' || line[nextText2] == '*')
		return true;
	return false;
}

/**
 * check if the currently reached '+' or '-' character is a unary operator
 * this method takes for granted that the current character
 * is a '+' or '-'.
 *
 * @return        whether the current '+' or '-' is a unary operator.
 */
bool ASFormatter::isUnaryOperator() const
{
	assert(currentChar == '+' || currentChar == '-');

	// does a digit follow a c-style cast
	if (previousCommandChar == ')')
	{
		if (!isdigit(peekNextChar()))
			return false;
		size_t end = currentLine.rfind(')', charNum);
		if (end == std::string::npos)
			return false;
		size_t lastChar = currentLine.find_last_not_of(" \t", end - 1);
		if (lastChar == std::string::npos)
			return false;
		if (currentLine[lastChar] == '*')
			end = lastChar;
		std::string prevWord = getPreviousWord(currentLine, end);
		if (prevWord.empty())
			return false;

		// a cast can be a user defined type
		//if (!isNumericVariable(prevWord))
		//	return false;

		return true;
	}

	return ((isCharImmediatelyPostReturn || !isLegalNameChar(previousCommandChar))
	        && previousCommandChar != '.'
	        && previousCommandChar != '\"'
	        && previousCommandChar != '\''
	        && previousCommandChar != ']');
}

/**
 * check if the currently reached comment is in a 'switch' statement
 *
 * @return        whether the current '+' or '-' is in an exponent.
 */
bool ASFormatter::isInSwitchStatement() const
{
	assert(isInLineComment || isInComment);
	if (!preBraceHeaderStack->empty())
		for (size_t i = 1; i < preBraceHeaderStack->size(); i++)
			if (preBraceHeaderStack->at(i) == &ASResource::AS_SWITCH)
				return true;
	return false;
}

/**
 * check if the currently reached '+' or '-' character is
 * part of an exponent, i.e. 0.2E-5.
 *
 * @return        whether the current '+' or '-' is in an exponent.
 */
bool ASFormatter::isInExponent() const
{
	assert(currentChar == '+' || currentChar == '-');
	std::string prevWord = getPreviousWord(currentLine, charNum, true);

	if (charNum && isDigit(prevWord[0]))
	{
		return prevWord.find_first_not_of("0123456789.") != std::string::npos;
	}

	if (charNum > 2 && prevWord.size() >= 2 && prevWord[0] == '0' && (prevWord[1] == 'x' || prevWord[1] == 'X'))
	{
		char prevPrevFormattedChar = currentLine[charNum - 2];
		char prevFormattedChar = currentLine[charNum - 1];
		//    double x = 0x1.23ffp-11;
		return ((prevFormattedChar == 'e' || prevFormattedChar == 'E' || prevFormattedChar == 'p' || prevFormattedChar == 'P')
		        && (prevPrevFormattedChar == '.' || std::isxdigit(prevPrevFormattedChar)));
	}
	return false;
}

/**
 * check if an array brace should NOT have an in-statement indent
 *
 * @return        the array is non in-statement
 */
bool ASFormatter::isNonInStatementArrayBrace() const
{
	bool returnVal = false;
	char nextChar = peekNextChar();
	// if this opening brace begins the line there will be no inStatement indent
	if (currentLineBeginsWithBrace
	        && (size_t) charNum == currentLineFirstBraceNum
	        && nextChar != '}')
		returnVal = true;
	// if an opening brace ends the line there will be no inStatement indent
	if (std::isblank(nextChar)
	        || isBeforeAnyLineEndComment(charNum)
	        || nextChar == '{')
		returnVal = true;

	// Java "new Type [] {...}" IS an inStatement indent
	if (isJavaStyle() && previousNonWSChar == ']')
		returnVal = false;

	return returnVal;
}

/**
 * check if a one-line block has been reached,
 * i.e. if the currently reached '{' character is closed
 * with a complimentary '}' elsewhere on the current line,
 *.
 * @return     0 = one-line block has not been reached.
 *             1 = one-line block has been reached.
 *             2 = one-line block has been reached and is followed by a comma.
 *             3 = one-line block has been reached and is an empty block.
 */
int ASFormatter::isOneLineBlockReached(std::string_view line, int startChar) const
{
	assert(line[startChar] == '{');

	bool isInComment_ = false;
	bool isInQuote_ = false;
	bool hasText = false;
	int braceCount = 0;
	int lineLength = line.length();
	char quoteChar_ = ' ';
	char ch = ' ';
	char prevCh = ' ';

	for (int i = startChar; i < lineLength; ++i)
	{
		ch = line[i];

		if (isInComment_)
		{
			if (line.compare(i, 2, "*/") == 0)
			{
				isInComment_ = false;
				++i;
			}
			continue;
		}

		if (isInQuote_)
		{
			if (ch == '\\')
				++i;
			else if (ch == quoteChar_)
				isInQuote_ = false;
			continue;
		}

		if (ch == '"'
		        || (ch == '\'' && !isDigitSeparator(line, i)))
		{
			isInQuote_ = true;
			quoteChar_ = ch;
			continue;
		}

		if (line.compare(i, 2, "//") == 0)
			break;

		if (line.compare(i, 2, "/*") == 0)
		{
			isInComment_ = true;
			++i;
			continue;
		}

		if (ch == '{')
		{
			++braceCount;
			continue;
		}
		if (ch == '}')
		{
			--braceCount;
			if (braceCount == 0)
			{
				// is this an array?
				if (parenStack->back() == 0 && prevCh != '}')
				{
					size_t peekNum = line.find_first_not_of(" \t", i + 1);
					if (peekNum != std::string::npos && line[peekNum] == ',')
						return 2;
				}
				if (!hasText)
					return 3;	// is an empty block
				return 1;
			}
		}
		if (ch == ';')
			continue;
		if (!std::isblank(ch))
		{
			hasText = true;
			prevCh = ch;
		}
	}

	return 0;
}

/**
 * peek at the next word to determine if it is a C# non-paren header.
 * will look ahead in the input file if necessary.
 *
 * @param  startChar      position on currentLine to start the search
 * @return                true if the next word is get or set.
 */
bool ASFormatter::isNextWordSharpNonParenHeader(int startChar) const
{
	// look ahead to find the next non-comment text
	std::string nextText = peekNextText(currentLine.substr(startChar));
	if (nextText.empty())
		return false;
	if (nextText[0] == '[')
		return true;
	if (!isCharPotentialHeader(nextText, 0))
		return false;
	if (findKeyword(nextText, 0, ASResource::AS_GET) || findKeyword(nextText, 0, ASResource::AS_SET)
	        || findKeyword(nextText, 0, ASResource::AS_ADD) || findKeyword(nextText, 0, ASResource::AS_REMOVE))
		return true;
	return false;
}

/**
 * peek at the next char to determine if it is an opening brace.
 * will look ahead in the input file if necessary.
 * this determines a java static constructor.
 *
 * @param startChar     position on currentLine to start the search
 * @return              true if the next word is an opening brace.
 */
bool ASFormatter::isNextCharOpeningBrace(int startChar) const
{
	bool retVal = false;
	std::string nextText = peekNextText(currentLine.substr(startChar));
	if (!nextText.empty()
	        && nextText.compare(0, 1, "{") == 0)
		retVal = true;
	return retVal;
}

/**
* Check if operator and, pointer, and reference padding is disabled.
* Disabling is done thru a NOPAD tag in an ending comment.
*
* @return              true if the formatting on this line is disabled.
*/
bool ASFormatter::isOperatorPaddingDisabled() const
{
	size_t commentStart = currentLine.find("//", charNum);
	if (commentStart == std::string::npos)
	{
		commentStart = currentLine.find("/*", charNum);
		// comment must end on this line
		if (commentStart != std::string::npos)
		{
			size_t commentEnd = currentLine.find("*/", commentStart + 2);
			if (commentEnd == std::string::npos)
				commentStart = std::string::npos;
		}
	}
	if (commentStart == std::string::npos)
		return false;
	size_t noPadStart = currentLine.find("*NOPAD*", commentStart);
	if (noPadStart == std::string::npos)
		return false;
	return true;
}

/**
* Determine if an opening array-type brace should have a leading space pad.
* This is to identify C++11 uniform initializers.
*/
bool ASFormatter::isUniformInitializerBrace() const
{
	if (isCStyle() && !isInEnum && !isImmediatelyPostPreprocessor)
	{
		if (isInClassInitializer
		        || isLegalNameChar(previousNonWSChar)
		        || previousNonWSChar == '(')
			return true;
	}
	return false;
}

/**
* Determine if there is a following statement on the current line.
*/
bool ASFormatter::isMultiStatementLine() const
{
	assert((isImmediatelyPostHeader || foundClosingHeader));
	bool isInComment_ = false;
	bool isInQuote_ = false;
	int  semiCount_ = 0;
	int  parenCount_ = 0;
	int  braceCount_ = 0;

	for (size_t i = 0; i < currentLine.length(); i++)
	{
		if (isInComment_)
		{
			if (currentLine.compare(i, 2, "*/") == 0)
			{
				isInComment_ = false;
				continue;
			}
		}
		if (currentLine.compare(i, 2, "/*") == 0)
		{
			isInComment_ = true;
			continue;
		}
		if (currentLine.compare(i, 2, "//") == 0)
			return false;
		if (isInQuote_)
		{
			if (currentLine[i] == '"' || currentLine[i] == '\'')
				isInQuote_ = false;
			continue;
		}
		if (currentLine[i] == '"' || currentLine[i] == '\'')
		{
			isInQuote_ = true;
			continue;
		}
		if (currentLine[i] == '(')
		{
			++parenCount_;
			continue;
		}
		if (currentLine[i] == ')')
		{
			--parenCount_;
			continue;
		}
		if (parenCount_ > 0)
			continue;
		if (currentLine[i] == '{')
		{
			++braceCount_;
		}
		if (currentLine[i] == '}')
		{
			--braceCount_;
		}
		if (braceCount_ > 0)
			continue;
		if (currentLine[i] == ';')
		{
			++semiCount_;
			if (semiCount_ > 1)
				return true;
			continue;
		}
	}
	return false;
}

/**
 * get the next non-whitespace substd::string on following lines, bypassing all comments.
 *
 * @param   firstLine   the first line to check
 * @return  the next non-whitespace substd::string.
 */
std::string ASFormatter::peekNextText(std::string_view firstLine,
                                      bool endOnEmptyLine /*false*/,
                                      const std::shared_ptr<ASPeekStream>& streamArg /*nullptr*/) const
{
	assert(sourceIterator->getPeekStart() == 0 || streamArg != nullptr);	// Borland may need != 0
	bool isFirstLine = true;
	std::string nextLine_(firstLine);
	size_t firstChar = std::string::npos;
	std::shared_ptr<ASPeekStream> stream = streamArg;
	if (stream == nullptr)					// Borland may need == 0
		stream = std::make_shared<ASPeekStream>(sourceIterator);

	// find the first non-blank text, bypassing all comments.
	bool isInComment_ = false;
	while (stream->hasMoreLines() || isFirstLine)
	{
		if (isFirstLine)
			isFirstLine = false;
		else
			nextLine_ = stream->peekNextLine();

		firstChar = nextLine_.find_first_not_of(" \t");
		if (firstChar == std::string::npos)
		{
			if (endOnEmptyLine && !isInComment_)
				break;
			continue;
		}

		if (nextLine_.compare(firstChar, 2, "/*") == 0)
		{
			firstChar += 2;
			isInComment_ = true;
		}

		if (isInComment_)
		{
			firstChar = nextLine_.find("*/", firstChar);
			if (firstChar == std::string::npos)
				continue;
			firstChar += 2;
			isInComment_ = false;
			firstChar = nextLine_.find_first_not_of(" \t", firstChar);
			if (firstChar == std::string::npos)
				continue;
		}

		if (nextLine_.compare(firstChar, 2, "//") == 0)
			continue;

		// found the next text
		break;
	}

	if (firstChar == std::string::npos)
		nextLine_ = "";
	else
		nextLine_ = nextLine_.substr(firstChar);
	return nextLine_;
}

/**
 * adjust comment position because of adding or deleting spaces
 * the spaces are added or deleted to formattedLine
 * spacePadNum contains the adjustment
 */
void ASFormatter::adjustComments()
{
	assert(spacePadNum != 0);
	assert(isSequenceReached(ASResource::AS_OPEN_LINE_COMMENT) || isSequenceReached(ASResource::AS_OPEN_COMMENT)  || isSequenceReached(ASResource::AS_GSC_OPEN_COMMENT));

	// block comment must be closed on this line with nothing after it
	bool isCppComment = isSequenceReached(ASResource::AS_OPEN_COMMENT);
	bool isGSCComment = isSequenceReached(ASResource::AS_GSC_OPEN_COMMENT);

	if (isCppComment || isGSCComment)
	{
		size_t endNum = currentLine.find(isCppComment ? ASResource::AS_CLOSE_COMMENT : ASResource::AS_GSC_CLOSE_COMMENT, charNum + 2);
		if (endNum == std::string::npos)
			return;
		// following line comments may be a tag from AStyleWx //[[)>
		size_t nextNum = currentLine.find_first_not_of(" \t", endNum + 2);
		if (nextNum != std::string::npos
		        && currentLine.compare(nextNum, 2, ASResource::AS_OPEN_LINE_COMMENT) != 0)
			return;
	}

	size_t len = formattedLine.length();
	// don't adjust a tab
	if (formattedLine[len - 1] == '\t')
		return;
	// if spaces were removed, need to add spaces before the comment
	if (spacePadNum < 0)
	{
		int adjust = -spacePadNum;          // make the number positive
		formattedLine.append(adjust, ' ');
	}
	// if spaces were added, need to delete extra spaces before the comment
	// if cannot be done put the comment one space after the last text
	else if (spacePadNum > 0)
	{
		int adjust = spacePadNum;
		size_t lastText = formattedLine.find_last_not_of(' ');
		if (lastText != std::string::npos
		        && lastText < len - adjust - 1)
			formattedLine.resize(len - adjust);
		else if (len > lastText + 2)
			formattedLine.resize(lastText + 2);
		else if (len < lastText + 2)
			formattedLine.append(len - lastText, ' ');
	}
}

/**
 * append the current brace inside the end of line comments
 * currentChar contains the brace, it will be appended to formattedLine
 * formattedLineCommentNum is the comment location on formattedLine
 */
void ASFormatter::appendCharInsideComments()
{
	if (formattedLineCommentNum == std::string::npos     // does the comment start on the previous line?
	        || formattedLineCommentNum == 0)
	{
		appendCurrentChar();                        // don't attach
		return;
	}
	assert(formattedLine.compare(formattedLineCommentNum, 2, "//") == 0
	       || formattedLine.compare(formattedLineCommentNum, 2, "/*") == 0);

	// find the previous non space char
	size_t end = formattedLineCommentNum;
	size_t beg = formattedLine.find_last_not_of(" \t", end - 1);
	if (beg == std::string::npos)
	{
		appendCurrentChar();                // don't attach
		return;
	}
	beg++;

	// insert the brace
	if (end - beg < 3)                      // is there room to insert?
		formattedLine.insert(beg, 3 - end + beg, ' ');
	if (formattedLine[beg] == '\t')         // don't pad with a tab
		formattedLine.insert(beg, 1, ' ');
	formattedLine[beg + 1] = currentChar;
	testForTimeToSplitFormattedLine();

	if (isBeforeComment())
		breakLine();
	else if (isCharImmediatelyPostLineComment)
		shouldBreakLineAtNextChar = true;
}

/**
 * add or remove space padding to operators
 * the operators and necessary padding will be appended to formattedLine
 * the calling function should have a continue statement after calling this method
 *
 * @param newOperator     the operator to be padded
 */
void ASFormatter::padOperators(const std::string* newOperator)
{
	assert(shouldPadOperators || negationPadMode != NEGATION_PAD_NO_CHANGE);
	assert(newOperator != nullptr);

	char nextNonWSChar = ASBase::peekNextChar(currentLine, charNum);
	std::set<char> allowedChars = {'(', '[', '=', ',', ':', '{'};

	bool isUnaryOrModOperator = (newOperator == &ASResource::AS_PLUS ||
	                             newOperator == &ASResource::AS_MINUS ||
	                             (newOperator == &ASResource::AS_MOD && isGSCStyle()));

	bool isExponentOperator = (newOperator == &ASResource::AS_MINUS && isInExponent()) ||
	                          (newOperator == &ASResource::AS_PLUS && isInExponent());

	bool isSpecialColon = (newOperator == &ASResource::AS_COLON && !foundQuestionMark &&
	                       (isInObjCMethodDefinition || isInObjCInterface || isInObjCSelector || squareBracketCount != 0));

	bool isJavaWildcard = (newOperator == &ASResource::AS_QUESTION && isJavaStyle() &&
	                       (previousNonWSChar == '<' || nextNonWSChar == '>' || nextNonWSChar == '.'));

	bool isSharpNullConditional = (newOperator == &ASResource::AS_QUESTION && isSharpStyle() &&
	                               (nextNonWSChar == '.' || nextNonWSChar == '['));

	bool isSpecialTemplateOperator = (isInTemplate || isImmediatelyPostTemplate) &&
	                                 (newOperator == &ASResource::AS_LS || newOperator == &ASResource::AS_GR);

	std::string sBegin = currentLine.substr(0, charNum);
	std::string sEnd = currentLine.substr(charNum, currentLine.find_first_not_of(">", charNum + 1));

	auto numOfOpeningBrackets = std::count(sBegin.begin(), sBegin.end(), '<');
	auto numOfClosingBrackets = std::count(sEnd.begin(), sEnd.end(), '>');

	bool isClosingTemplateDefinition = numOfClosingBrackets >= numOfOpeningBrackets && numOfOpeningBrackets >= 2;

	bool shouldPad = (newOperator != &ASResource::AS_SCOPE_RESOLUTION &&
	                  newOperator != &ASResource::AS_PLUS_PLUS &&
	                  newOperator != &ASResource::AS_MINUS_MINUS &&
	                  (newOperator != &ASResource::AS_NOT || negationPadMode != NEGATION_PAD_NO_CHANGE) &&
	                  newOperator != &ASResource::AS_BIT_NOT &&
	                  newOperator != &ASResource::AS_ARROW &&
	                  !isSpecialColon &&
	                  !isExponentOperator &&
	                  !isClosingTemplateDefinition &&
	                  !(newOperator == &ASResource::AS_GR && previousChar == '-') &&
	                  !(isUnaryOrModOperator && (allowedChars.find(previousNonWSChar) != allowedChars.end())) &&
	                  !(newOperator == &ASResource::AS_MULT &&
	                    (previousNonWSChar == '.' || previousNonWSChar == '>')) &&
	                  !(newOperator == &ASResource::AS_MULT && peekNextChar() == '>') &&
	                  !isSpecialTemplateOperator &&
	                  !(newOperator == &ASResource::AS_GCC_MIN_ASSIGN &&
	                    ASBase::peekNextChar(currentLine, charNum + 1) == '>') &&
	                  !(newOperator == &ASResource::AS_GR && previousNonWSChar == '?') &&
	                  !isJavaWildcard &&
	                  !isSharpNullConditional &&
	                  !isCharImmediatelyPostOperator &&
	                  !isInCase &&
	                  !isInAsm &&
	                  !isInAsmOneLine &&
	                  !isInAsmBlock);

	// pad before operator
	if (shouldPad
	        && (newOperator != &ASResource::AS_NOT || (newOperator == &ASResource::AS_NOT && negationPadMode == NEGATION_PAD_BEFORE ) )
	        && !(newOperator == &ASResource::AS_COLON
	             && (!foundQuestionMark && !isInEnum) && currentHeader != &ASResource::AS_FOR)
	        && !(newOperator == &ASResource::AS_QUESTION && isSharpStyle() // check for C# nullable type (e.g. int?)
	             && currentLine.find(':', charNum + 1) == std::string::npos)
	   )
	{
		appendSpacePad();
	}

	appendOperator(*newOperator);
	goForward(newOperator->length() - 1);

	currentChar = (*newOperator)[newOperator->length() - 1];
	// pad after operator
	// but do not pad after a '-' that is a unary-minus.
	if (shouldPad
	        && !isBeforeAnyComment()
	        && !(newOperator == &ASResource::AS_PLUS && isUnaryOperator())
	        && !(newOperator == &ASResource::AS_MINUS && isUnaryOperator())
	        && !(currentLine.compare(charNum + 1, 1, ASResource::AS_SEMICOLON) == 0)
	        && !(currentLine.compare(charNum + 1, 2, ASResource::AS_SCOPE_RESOLUTION) == 0)
	        && !(peekNextChar() == ',')
	        && !(newOperator == &ASResource::AS_QUESTION && isSharpStyle() // check for C# nullable type (e.g. int?)
	             && peekNextChar() == '[')
	   )
	{
		appendSpaceAfter();
	}
}

/**
 * format pointer or reference
 * currentChar contains the pointer or reference
 * the symbol and necessary padding will be appended to formattedLine
 * the calling function should have a continue statement after calling this method
 *
 * NOTE: Do NOT use appendCurrentChar() in this method. The line should not be
 *       broken once the calculation starts.
 */
void ASFormatter::formatPointerOrReference()
{
	assert(currentChar == '*' || currentChar == '&' || currentChar == '^');
	assert(!isJavaStyle());

	int pa = pointerAlignment;
	int ra = referenceAlignment;
	int itemAlignment = (currentChar == '*' || currentChar == '^')
	                    ? pa : ((ra == REF_SAME_AS_PTR) ? pa : ra);

	// handle operator char*() {};
	if (currentChar == '*' && isCStyle() && currentLine.find("operator") != std::string::npos)
	{
		formattedLine.append("*");
		return;
	}

	// check for ** and &&
	int ptrLength = 1;
	char peekedChar = peekNextChar();
	if ((currentChar == '*' && peekedChar == '*')
	        || (currentChar == '&' && peekedChar == '&'))
	{
		ptrLength = 2;

		//TODO check
		size_t nextChar = currentLine.find_first_not_of(" \t", charNum + 2);
		if (nextChar == std::string::npos)
			peekedChar = ' ';
		else
			peekedChar = currentLine[nextChar];

		//https://sourceforge.net/p/astyle/bugs/543/
		if (currentChar == '&' /*&& itemAlignment == PTR_ALIGN_NAME*/)
		{
			itemAlignment = PTR_ALIGN_NONE;
		}
	}
	// check for cast
	if (peekedChar == ')' || peekedChar == '>' || peekedChar == ',')
	{
		formatPointerOrReferenceCast();
		return;
	}

	// check for a padded space and remove it
	if (charNum > 0
	        && !std::isblank(currentLine[charNum - 1])
	        && !formattedLine.empty()
	        && std::isblank(formattedLine[formattedLine.length() - 1]))
	{
		formattedLine.erase(formattedLine.length() - 1);
		spacePadNum--;
	}

	if (itemAlignment == PTR_ALIGN_TYPE)
	{
		formatPointerOrReferenceToType();
	}
	else if (itemAlignment == PTR_ALIGN_MIDDLE)
	{
		formatPointerOrReferenceToMiddle();
	}
	else if (itemAlignment == PTR_ALIGN_NAME)
	{
		formatPointerOrReferenceToName();
	}
	else	// pointerAlignment == PTR_ALIGN_NONE
	{
		formattedLine.append(currentLine.substr(charNum, ptrLength));
		if (ptrLength > 1)
			goForward(ptrLength - 1);
	}
}

/**
 * format pointer or reference with align to type
 */
void ASFormatter::formatPointerOrReferenceToType()
{
	assert(currentChar == '*' || currentChar == '&' || currentChar == '^');
	assert(!isJavaStyle());

	// do this before bumping charNum
	bool isOldPRCentered = isPointerOrReferenceCentered();
	std::string sequenceToInsert(1, currentChar);
	// get the sequence
	if (currentChar == peekNextChar())
	{
		for (size_t i = charNum + 1; currentLine.length() > i; i++)
		{
			if (currentLine[i] == sequenceToInsert[0])
			{
				sequenceToInsert.append(1, currentLine[i]);
				goForward(1);
				continue;
			}
			break;
		}
	}
	// append the sequence
	std::string charSave;
	size_t prevCh = formattedLine.find_last_not_of(" \t");
	if (prevCh < formattedLine.length())
	{
		charSave = formattedLine.substr(prevCh + 1);
		formattedLine.resize(prevCh + 1);
	}

	// https://sourceforge.net/p/astyle/bugs/537/
	// TODO check
	if ((previousNonWSChar == ',' || previousNonWSChar == '[') && currentChar != ' ')
		appendSpacePad();

	formattedLine.append(sequenceToInsert);
	if (peekNextChar() != ')')
		formattedLine.append(charSave);
	else
		spacePadNum -= charSave.length();
	// if no space after then add one
	if (charNum < (int) currentLine.length() - 1
	        && !std::isblank(currentLine[charNum + 1])
	        && currentLine[charNum + 1] != ')'
		    && peekNextChar() != '&')
		appendSpacePad();

	// if old pointer or reference is centered, remove a space
	if (isOldPRCentered
	        && std::isblank(formattedLine[formattedLine.length() - 1]))
	{
		formattedLine.erase(formattedLine.length() - 1, 1);
		spacePadNum--;
	}
	// update the formattedLine split point
	if (maxCodeLength != std::string::npos && !formattedLine.empty())
	{
		size_t index = formattedLine.length() - 1;
		if (std::isblank(formattedLine[index]))
		{
			updateFormattedLineSplitPointsPointerOrReference(index);
			testForTimeToSplitFormattedLine();
		}
	}
}

/**
 * format pointer or reference with align in the middle
 */
void ASFormatter::formatPointerOrReferenceToMiddle()
{
	assert(currentChar == '*' || currentChar == '&' || currentChar == '^');
	assert(!isJavaStyle());

	if (currentLine.size() > (size_t)charNum + 1 &&
	        std::isblank(currentLine[charNum - 1]) &&
	        std::isblank(currentLine[charNum + 1]))
	{

		std::string seq = {currentChar, currentLine[charNum + 1]};
		appendSequence(seq);
		goForward(1);
		return;
	}

	// compute current whitespace before
	size_t wsBefore = currentLine.find_last_not_of(" \t", charNum - 1);
	if (wsBefore == std::string::npos)
		wsBefore = 0;
	else
		wsBefore = charNum - wsBefore - 1;
	std::string sequenceToInsert(1, currentChar);

	if (currentChar == peekNextChar())
	{
		for (size_t i = charNum + 1; currentLine.length() > i; i++)
		{
			if (currentLine[i] == sequenceToInsert[0])
			{
				sequenceToInsert.append(1, currentLine[i]);
				goForward(1);
				continue;
			}
			break;
		}
	}
	// if reference to a pointer check for conflicting alignment
	else if (currentChar == '*' && peekNextChar() == '&'
	         && ASBeautifier::peekNextChar(currentLine, charNum + 1) != '&'
	         && (referenceAlignment == REF_ALIGN_TYPE
	             || referenceAlignment == REF_ALIGN_MIDDLE
	             || referenceAlignment == REF_SAME_AS_PTR))
	{
		sequenceToInsert = "*&";
		goForward(1);
		for (size_t i = charNum; i < currentLine.length() - 1 && std::isblank(currentLine[i]); i++)
			goForward(1);
	}
	// if a comment follows don't align, just space pad
	if (isBeforeAnyComment())
	{
		appendSpacePad();
		formattedLine.append(sequenceToInsert);
		appendSpaceAfter();
		return;
	}
	// do this before goForward()
	bool isAfterScopeResolution = previousNonWSChar == ':';
	size_t charNumSave = charNum;
	// if this is the last thing on the line
	if (currentLine.find_first_not_of(" \t", charNum + 1) == std::string::npos)
	{
		if (wsBefore == 0 && !isAfterScopeResolution)
			formattedLine.append(1, ' ');
		formattedLine.append(sequenceToInsert);
		return;
	}
	// goForward() to convert tabs to spaces, if necessary,
	// and move following characters to preceding characters
	// this may not work every time with tab characters
	for (size_t i = charNum + 1; i < currentLine.length() && std::isblank(currentLine[i]); i++)
	{
		goForward(1);
		if (!formattedLine.empty())
			formattedLine.append(1, currentLine[i]);
		else
			spacePadNum--;
	}
	// find space padding after
	size_t wsAfter = currentLine.find_first_not_of(" \t", charNumSave + 1);
	if (wsAfter == std::string::npos || isBeforeAnyComment())
		wsAfter = 0;
	else
		wsAfter = wsAfter - charNumSave - 1;
	// don't pad before scope resolution operator, but pad after
	if (isAfterScopeResolution)
	{
		size_t lastText = formattedLine.find_last_not_of(" \t");
		formattedLine.insert(lastText + 1, sequenceToInsert);
		appendSpacePad();
	}
	else if (!formattedLine.empty())
	{
		// whitespace should be at least 2 chars to center
		if (wsBefore + wsAfter < 2)
		{
			size_t charsToAppend = (2 - (wsBefore + wsAfter));
			formattedLine.append(charsToAppend, ' ');
			spacePadNum += charsToAppend;
			if (wsBefore == 0)
				wsBefore++;
			if (wsAfter == 0)
				wsAfter++;
		}
		// insert the pointer or reference char
		size_t padAfter = (wsBefore + wsAfter) / 2;
		size_t index = formattedLine.length() - padAfter;
		if (index < formattedLine.length())
			formattedLine.insert(index, sequenceToInsert);
		else
			formattedLine.append(sequenceToInsert);
	}
	else	// formattedLine.length() == 0
	{
		formattedLine.append(sequenceToInsert);
		if (wsAfter == 0)
			wsAfter++;
		formattedLine.append(wsAfter, ' ');
		spacePadNum += wsAfter;
	}
	// update the formattedLine split point after the pointer
	if (maxCodeLength != std::string::npos && !formattedLine.empty())
	{
		size_t index = formattedLine.find_last_not_of(" \t");
		if (index != std::string::npos && (index < formattedLine.length() - 1))
		{
			index++;
			updateFormattedLineSplitPointsPointerOrReference(index);
			testForTimeToSplitFormattedLine();
		}
	}
}

/**
 * format pointer or reference with align to name
 */
void ASFormatter::formatPointerOrReferenceToName()
{
	assert(currentChar == '*' || currentChar == '&' || currentChar == '^');
	assert(!isJavaStyle());

	// do this before bumping charNum
	bool isOldPRCentered = isPointerOrReferenceCentered();
	size_t startNum = formattedLine.find_last_not_of(" \t");
	if (startNum == std::string::npos)
		startNum = 0;
	std::string sequenceToInsert(1, currentChar);
	if (currentChar == peekNextChar())
	{
		for (size_t i = charNum + 1; currentLine.length() > i; i++)
		{
			if (currentLine[i] == sequenceToInsert[0])
			{
				sequenceToInsert.append(1, currentLine[i]);
				goForward(1);
				continue;
			}
			break;
		}
	}

	// if reference to a pointer align both to name

	else if (currentChar == '*' && peekNextChar() == '&' && ASBeautifier::peekNextChar(currentLine, charNum + 1) != '&')
	{
		sequenceToInsert = "*&";
		goForward(1);
		for (size_t i = charNum; i < currentLine.length() - 1 && std::isblank(currentLine[i]); i++)
			goForward(1);
	}


	char peekedChar = peekNextChar();
	bool isAfterScopeResolution = previousNonWSChar == ':';		// check for ::
	// if this is not the last thing on the line
	if ((isLegalNameChar(peekedChar) || peekedChar == '(' || peekedChar == '[' || peekedChar == '=')
	        && (int) currentLine.find_first_not_of(" \t", charNum + 1) > charNum)
	{
		// goForward() to convert tabs to spaces, if necessary,
		// and move following characters to preceding characters
		// this may not work every time with tab characters
		for (size_t i = charNum + 1; i < currentLine.length() && std::isblank(currentLine[i]); i++)
		{
			// if a padded paren follows don't move
			if (shouldPadParensOutside && peekedChar == '(' && !isOldPRCentered)
			{
				// empty parens don't count
				size_t start = currentLine.find_first_not_of("( \t", i);
				if (start != std::string::npos && currentLine[start] != ')')
					break;
			}
			goForward(1);
			if (!formattedLine.empty())
				formattedLine.append(1, currentLine[charNum]);
			else
				spacePadNum--;
		}
	}
	// don't pad before scope resolution operator
	if (isAfterScopeResolution)
	{
		size_t lastText = formattedLine.find_last_not_of(" \t");
		if (lastText != std::string::npos && lastText + 1 < formattedLine.length())
			formattedLine.erase(lastText + 1);
	}
	// if no space before * then add one
	else if (!formattedLine.empty()
			&& (currentLine[startNum + 1] != '&')
			&& (formattedLine.length() <= startNum + 1
				|| !std::isblank(formattedLine[startNum + 1])))
	{
		formattedLine.insert(startNum + 1, 1, ' ');
		spacePadNum++;
	}
	appendSequence(sequenceToInsert, false);

	// if old pointer or reference is centered, remove a space
	if (isOldPRCentered
	        && formattedLine.length() > startNum + 1
	        && std::isblank(formattedLine[startNum + 1])
	        && peekedChar != '*'		// check for '* *'
	        && !isBeforeAnyComment()
	        && ((isLegalNameChar(peekedChar) || peekedChar == '(') && pointerAlignment == PTR_ALIGN_NAME) //https://sourceforge.net/p/astyle/bugs/546/ + #527
	   )
	{
		formattedLine.erase(startNum + 1, 1);
		spacePadNum--;
	}
	// don't convert to *= or &=
	if (peekedChar == '=')
	{
		appendSpaceAfter();
		// if more than one space before, delete one
		if (formattedLine.length() > startNum
		        && std::isblank(formattedLine[startNum + 1])
		        && std::isblank(formattedLine[startNum + 2]))
		{
			formattedLine.erase(startNum + 1, 1);
			spacePadNum--;
		}
	}
	// update the formattedLine split point
	if (maxCodeLength != std::string::npos)
	{
		size_t index = formattedLine.find_last_of(" \t");
		if (index != std::string::npos
		        && index < formattedLine.length() - 1
		        && (formattedLine[index + 1] == '*'
		            || formattedLine[index + 1] == '&'
		            || formattedLine[index + 1] == '^'))
		{
			updateFormattedLineSplitPointsPointerOrReference(index);
			testForTimeToSplitFormattedLine();
		}
	}
}

/**
 * format pointer or reference cast
 * currentChar contains the pointer or reference
 * NOTE: the pointers and references in function definitions
 *       are processed as a cast (e.g. void foo(void*, void*))
 *       is processed here.
 */
void ASFormatter::formatPointerOrReferenceCast()
{
	assert(currentChar == '*' || currentChar == '&' || currentChar == '^');
	assert(!isJavaStyle());

	int pa = pointerAlignment;
	int ra = referenceAlignment;
	int itemAlignment = (currentChar == '*' || currentChar == '^')
	                    ? pa : ((ra == REF_SAME_AS_PTR) ? pa : ra);

	std::string sequenceToInsert(1, currentChar);
	if (isSequenceReached("**") || isSequenceReached(ASResource::AS_AND))
	{
		goForward(1);
		sequenceToInsert.append(1, currentLine[charNum]);
	}
	if (itemAlignment == PTR_ALIGN_NONE)
	{
		appendSequence(sequenceToInsert, false);
		return;
	}
	// remove preceding whitespace
	char prevCh = ' ';
	size_t prevNum = formattedLine.find_last_not_of(" \t");
	if (prevNum != std::string::npos)
	{
		prevCh = formattedLine[prevNum];
		if (itemAlignment == PTR_ALIGN_TYPE && currentChar == '*' && prevCh == '*')
		{
			// '* *' may be a multiply followed by a dereference
			if (prevNum + 2 < formattedLine.length()
			        && std::isblank(formattedLine[prevNum + 2]))
			{
				spacePadNum -= (formattedLine.length() - 2 - prevNum);
				formattedLine.erase(prevNum + 2);
			}
		}
		else if (prevNum + 1 < formattedLine.length()
		         && std::isblank(formattedLine[prevNum + 1])
		         && prevCh != '(')
		{
			spacePadNum -= (formattedLine.length() - 1 - prevNum);
			formattedLine.erase(prevNum + 1);
		}
	}

	bool isAfterScopeResolution = previousNonWSChar == ':';
	if ((itemAlignment == PTR_ALIGN_MIDDLE || itemAlignment == PTR_ALIGN_NAME)
	        && !isAfterScopeResolution && prevCh != '(')
	{
		appendSpacePad();
		// in this case appendSpacePad may or may not update the split point
		if (maxCodeLength != std::string::npos && !formattedLine.empty())
			updateFormattedLineSplitPointsPointerOrReference(formattedLine.length() - 1);
		appendSequence(sequenceToInsert, false);
	}
	else
		appendSequence(sequenceToInsert, false);
}

/**
 * add or remove space padding to parens
 * currentChar contains the paren
 * the parens and necessary padding will be appended to formattedLine
 * the calling function should have a continue statement after calling this method
 */
void ASFormatter::padParensOrBrackets(char openDelim, char closeDelim, bool padFirstParen)
{
	assert(currentChar == openDelim || currentChar == closeDelim);

	int spacesOutsideToDelete = 0;
	int spacesInsideToDelete = 0;

	bool shouldPadOutside = shouldPadParensOutside || shouldPadBracketsOutside;
	bool shouldPadInside = shouldPadParensInside || shouldPadBracketsInside;
	bool shouldUnPad = shouldUnPadParens || shouldUnPadBrackets;

	if (currentChar == openDelim)
	{
		spacesOutsideToDelete = formattedLine.length() - 1;
		spacesInsideToDelete = 0;

		// compute spaces outside the opening paren to delete
		if (shouldUnPad && !isInStruct)
		{
			char lastChar = ' ';
			bool prevIsParenHeader = false;
			size_t i = formattedLine.find_last_not_of(" \t");
			if (i != std::string::npos)
			{
				// if last char is a brace the previous whitespace is an indent
				if (formattedLine[i] == '{')
					spacesOutsideToDelete = 0;
				else if (isCharImmediatelyPostPointerOrReference)
					spacesOutsideToDelete = 0;
				else
				{
					spacesOutsideToDelete -= i;
					lastChar = formattedLine[i];
					// if previous word is a header, it will be a paren header
					std::string prevWord = getPreviousWord(formattedLine, formattedLine.length());
					const std::string* prevWordH = nullptr;
					if (shouldPadHeader
					        && !prevWord.empty()
					        && isCharPotentialHeader(prevWord, 0))
						prevWordH = ASBase::findHeader(prevWord, 0, headers);

					if (prevWordH != nullptr)
						prevIsParenHeader = true;    // don't unpad
					else if (prevWord == ASResource::AS_RETURN)
						prevIsParenHeader = true;    // don't unpad
					else if ((prevWord == ASResource::AS_NEW || prevWord == ASResource::AS_DELETE)
					         && shouldPadHeader)
						prevIsParenHeader = true;    // don't unpad
					else if (isCStyle() && prevWord == ASResource::AS_THROW && shouldPadHeader)
						prevIsParenHeader = true;    // don't unpad
					else if (prevWord == "and" || prevWord == "or" || prevWord == "in")
						prevIsParenHeader = true;    // don't unpad
					// don't unpad variables
					else if (isNumericVariable(prevWord))
						prevIsParenHeader = true;    // don't unpad
				}
			}
			// do not unpad operators, but leave them if already padded
			if (shouldPadOutside || prevIsParenHeader)
			{
				spacesOutsideToDelete--;
			}
			else
			{
				static const std::string operatorList = "|&<>,?:;=+-*/%^";
				if (operatorList.find(lastChar) != std::string::npos ||
				        (lastChar == openDelim && shouldPadInside) ||
				        (lastChar == '>' && !foundCastOperator))
				{
					spacesOutsideToDelete--;
				}
			}

			if (spacesOutsideToDelete > 0)
			{
				formattedLine.erase(i + 1, spacesOutsideToDelete);
				spacePadNum -= spacesOutsideToDelete;
			}
		}

		// pad open paren outside
		char peekedCharOutside = peekNextChar();
		if (padFirstParen && ( (previousChar != openDelim && peekedCharOutside != closeDelim)  || shouldPadEmptyParens ) )
			appendSpacePad();
		else if (shouldPadOutside)
		{
			// GH19
			if (!(currentChar == openDelim && peekedCharOutside == closeDelim) || shouldPadEmptyParens)
				appendSpacePad();
		}

		appendCurrentChar();

		// unpad open paren inside
		if (shouldUnPad)
		{
			size_t j = currentLine.find_first_not_of(" \t", charNum + 1);
			if (j != std::string::npos)
				spacesInsideToDelete = j - charNum - 1;
			if (shouldPadInside)
				spacesInsideToDelete--;
			if (spacesInsideToDelete > 0)
			{
				currentLine.erase(charNum + 1, spacesInsideToDelete);
				spacePadNum -= spacesInsideToDelete;
			}
			// convert tab to space if requested
			if (shouldConvertTabs
			        && (int) currentLine.length() > charNum + 1
			        && currentLine[charNum + 1] == '\t')
				currentLine[charNum + 1] = ' ';
		}

		// pad open paren inside
		char peekedCharInside = peekNextChar();
		if (shouldPadInside)
			if (!(currentChar == openDelim && peekedCharInside == closeDelim))
				appendSpaceAfter();
	}
	else if (currentChar == closeDelim)
	{
		// unpad close paren inside
		if (shouldUnPad)
		{
			spacesInsideToDelete = formattedLine.length();
			size_t i = formattedLine.find_last_not_of(" \t");
			if (i != std::string::npos)
				spacesInsideToDelete = formattedLine.length() - 1 - i;
			if (shouldPadInside)
				spacesInsideToDelete--;
			if (spacesInsideToDelete > 0)
			{
				formattedLine.erase(i + 1, spacesInsideToDelete);
				spacePadNum -= spacesInsideToDelete;
			}
		}

		// pad close paren inside
		if (shouldPadInside)
			if (!(previousChar == openDelim && currentChar == closeDelim))
				appendSpacePad();

		appendCurrentChar();

		// pad close paren outside
		char peekedCharOutside = peekNextChar();
		if (shouldPadOutside)
			if (peekedCharOutside != ';'
			        && peekedCharOutside != ','
			        && peekedCharOutside != '.'
			        && peekedCharOutside != '+'    // check for ++
			        && peekedCharOutside != '-'    // check for --
			        && peekedCharOutside != ']')
			{
				appendSpaceAfter();
			}
	}
}

/**
* add or remove space padding to objective-c method prefix (- or +)
* if this is a '(' it begins a return type
* these options have precedence over the padParensOrBrackets methods
* the padParensOrBrackets method has already been called, this method adjusts
*/
void ASFormatter::padObjCMethodPrefix()
{
	assert(isInObjCMethodDefinition && isImmediatelyPostObjCMethodPrefix);
	assert(shouldPadMethodPrefix || shouldUnPadMethodPrefix);

	size_t prefix = formattedLine.find_first_of("+-");
	if (prefix == std::string::npos)
		return;
	size_t firstChar = formattedLine.find_first_not_of(" \t", prefix + 1);
	if (firstChar == std::string::npos)
		firstChar = formattedLine.length();
	int spaces = firstChar - prefix - 1;

	if (shouldPadMethodPrefix)
	{
		if (spaces == 0)
		{
			formattedLine.insert(prefix + 1, 1, ' ');
			spacePadNum += 1;
		}
		else if (spaces > 1)
		{
			formattedLine.erase(prefix + 1, spaces - 1);
			formattedLine[prefix + 1] = ' ';  // convert any tab to space
			spacePadNum -= spaces - 1;
		}
	}
	// this option will be ignored if used with pad-method-prefix
	else if (shouldUnPadMethodPrefix)
	{
		if (spaces > 0)
		{
			formattedLine.erase(prefix + 1, spaces);
			spacePadNum -= spaces;
		}
	}
}

/**
* add or remove space padding to objective-c parens
* these options have precedence over the padParensOrBrackets methods
* the padParensOrBrackets method has already been called, this method adjusts
*/
void ASFormatter::padObjCReturnType()
{
	assert(currentChar == ')' && isInObjCReturnType);
	assert(shouldPadReturnType || shouldUnPadReturnType);

	size_t nextText = currentLine.find_first_not_of(" \t", charNum + 1);
	if (nextText == std::string::npos)
		return;
	int spaces = nextText - charNum - 1;

	if (shouldPadReturnType)
	{
		if (spaces == 0)
		{
			// this will already be padded if pad-paren is used
			if (formattedLine[formattedLine.length() - 1] != ' ')
			{
				formattedLine.append(" ");
				spacePadNum += 1;
			}
		}
		else if (spaces > 1)
		{
			// do not use goForward here
			currentLine.erase(charNum + 1, spaces - 1);
			currentLine[charNum + 1] = ' ';  // convert any tab to space
			spacePadNum -= spaces - 1;
		}
	}
	// this option will be ignored if used with pad-return-type
	else if (shouldUnPadReturnType)
	{
		// this will already be padded if pad-paren is used
		if (formattedLine[formattedLine.length() - 1] == ' ')
		{
			int lastText = formattedLine.find_last_not_of(" \t");
			spacePadNum -= formattedLine.length() - lastText - 1;
			formattedLine.resize(lastText + 1);
		}
		// do not use goForward here
		currentLine.erase(charNum + 1, spaces);
		spacePadNum -= spaces;
	}
}

/**
* add or remove space padding to objective-c parens
* these options have precedence over the padParensOrBrackets methods
* the padParensOrBrackets method has already been called, this method adjusts
*/
void ASFormatter::padObjCParamType()
{
	assert((currentChar == '(' || currentChar == ')') && isInObjCMethodDefinition);
	assert(!isImmediatelyPostObjCMethodPrefix && !isInObjCReturnType);
	assert(shouldPadParamType || shouldUnPadParamType);

	if (currentChar == '(')
	{
		// open paren has already been attached to formattedLine by padParen
		size_t paramOpen = formattedLine.rfind('(');
		assert(paramOpen != std::string::npos);
		size_t prevText = formattedLine.find_last_not_of(" \t", paramOpen - 1);
		if (prevText == std::string::npos)
			return;
		int spaces = paramOpen - prevText - 1;

		if (shouldPadParamType
		        || objCColonPadMode == COLON_PAD_ALL
		        || objCColonPadMode == COLON_PAD_AFTER)
		{
			if (spaces == 0)
			{
				formattedLine.insert(paramOpen, 1, ' ');
				spacePadNum += 1;
			}
			if (spaces > 1)
			{
				formattedLine.erase(prevText + 1, spaces - 1);
				formattedLine[prevText + 1] = ' ';  // convert any tab to space
				spacePadNum -= spaces - 1;
			}
		}
		// this option will be ignored if used with pad-param-type
		else if (shouldUnPadParamType
		         || objCColonPadMode == COLON_PAD_NONE
		         || objCColonPadMode == COLON_PAD_BEFORE)
		{
			if (spaces > 0)
			{
				formattedLine.erase(prevText + 1, spaces);
				spacePadNum -= spaces;
			}
		}
	}
	else if (currentChar == ')')
	{
		size_t nextText = currentLine.find_first_not_of(" \t", charNum + 1);
		if (nextText == std::string::npos)
			return;
		int spaces = nextText - charNum - 1;

		if (shouldPadParamType)
		{
			if (spaces == 0)
			{
				// this will already be padded if pad-paren is used
				if (formattedLine[formattedLine.length() - 1] != ' ')
				{
					formattedLine.append(" ");
					spacePadNum += 1;
				}
			}
			else if (spaces > 1)
			{
				// do not use goForward here
				currentLine.erase(charNum + 1, spaces - 1);
				currentLine[charNum + 1] = ' ';  // convert any tab to space
				spacePadNum -= spaces - 1;
			}
		}
		// this option will be ignored if used with pad-param-type
		else if (shouldUnPadParamType)
		{
			// this will already be padded if pad-paren is used
			if (formattedLine[formattedLine.length() - 1] == ' ')
			{
				spacePadNum -= 1;
				int lastText = formattedLine.find_last_not_of(" \t");
				formattedLine.resize(lastText + 1);
			}
			if (spaces > 0)
			{
				// do not use goForward here
				currentLine.erase(charNum + 1, spaces);
				spacePadNum -= spaces;
			}
		}
	}
}

/**
 * format opening brace as attached or broken
 * currentChar contains the brace
 * the braces will be appended to the current formattedLine or a new formattedLine as necessary
 * the calling function should have a continue statement after calling this method
 *
 * @param braceType    the type of brace to be formatted.
 */
void ASFormatter::formatOpeningBrace(BraceType braceType)
{
	assert(!isBraceType(braceType, ARRAY_TYPE));
	assert(currentChar == '{');

	parenStack->emplace_back(0);

	bool breakBrace = isCurrentBraceBroken();

	//478

	if (breakBrace)
	{
		if (isBeforeAnyComment() && isOkToBreakBlock(braceType) && sourceIterator->hasMoreLines())
		{
			// if comment is at line end leave the comment on this line
			if (isBeforeAnyLineEndComment(charNum) && !currentLineBeginsWithBrace)
			{
				currentChar = ' ';              // remove brace from current line
				if (parenStack->size() > 1)
					parenStack->pop_back();
				currentLine[charNum] = currentChar;
				appendOpeningBrace = true;      // append brace to following line
			}
			// else put comment after the brace
			else if (!isBeforeMultipleLineEndComments(charNum))
				breakLine();
		}
		else if (!isBraceType(braceType, SINGLE_LINE_TYPE))
		{
			formattedLine = rtrim(formattedLine);
			breakLine();
		}
		else if ((shouldBreakOneLineBlocks || isBraceType(braceType, BREAK_BLOCK_TYPE))
		         && !isBraceType(braceType, EMPTY_BLOCK_TYPE))
			breakLine();
		else if (!isInLineBreak)
			appendSpacePad();

		appendCurrentChar();

		// should a following comment break from the brace?
		// must break the line AFTER the brace
		if (isBeforeComment()
		        && !formattedLine.empty()
		        && formattedLine[0] == '{'
		        && isOkToBreakBlock(braceType)
		        && (braceFormatMode == BREAK_MODE
		            || braceFormatMode == LINUX_MODE))
		{
			shouldBreakLineAtNextChar = true;
		}
	}
	else    // attach brace
	{
		// are there comments before the brace?
		if (isCharImmediatelyPostComment || isCharImmediatelyPostLineComment)
		{
			if (isOkToBreakBlock(braceType)
			        && !(isCharImmediatelyPostComment && isCharImmediatelyPostLineComment)	// don't attach if two comments on the line
			        && !isImmediatelyPostPreprocessor
			        && previousCommandChar != '{'	// don't attach { {
			        && previousCommandChar != '}'	// don't attach } {
			        && previousCommandChar != ';')	// don't attach ; {
			{
				appendCharInsideComments();
			}
			else
			{
				appendCurrentChar();				// don't attach
			}
		}
		else if (previousCommandChar == '{'
		         || (previousCommandChar == '}' && !isInClassInitializer)
		         || previousCommandChar == ';')		// '}' , ';' chars added for proper handling of '{' immediately after a '}' or ';'
		{
			appendCurrentChar();					// don't attach
		}
		else
		{
			// if a blank line precedes this don't attach
			if (isEmptyLine(formattedLine))
				appendCurrentChar();				// don't attach
			else if (isOkToBreakBlock(braceType)
			         && !(isImmediatelyPostPreprocessor
			              && currentLineBeginsWithBrace))
			{
				if (!isBraceType(braceType, EMPTY_BLOCK_TYPE))
				{
					appendSpacePad();
					appendCurrentChar(false);				// OK to attach
					testForTimeToSplitFormattedLine();		// line length will have changed
					// should a following comment attach with the brace?
					// insert spaces to reposition the comment
					if (isBeforeComment()
					        && !isBeforeMultipleLineEndComments(charNum)
					        && (!isBeforeAnyLineEndComment(charNum)	|| currentLineBeginsWithBrace))
					{
						shouldBreakLineAtNextChar = true;
						currentLine.insert(charNum + 1, charNum + 1, ' ');
					}
					else if (!isBeforeAnyComment())		// added in release 2.03
					{
						shouldBreakLineAtNextChar = true;
					}
				}
				else
				{
					if (currentLineBeginsWithBrace && (size_t) charNum == currentLineFirstBraceNum)
					{
						appendSpacePad();
						appendCurrentChar(false);		// attach
						shouldBreakLineAtNextChar = true;
					}
					else
					{
						appendSpacePad();
						appendCurrentChar();		// don't attach
					}
				}
			}
			else
			{
				if (!isInLineBreak)
					appendSpacePad();
				appendCurrentChar();				// don't attach
			}
		}
	}
}

/**
 * format closing brace
 * currentChar contains the brace
 * the calling function should have a continue statement after calling this method
 *
 * @param braceType    the type of the opening brace for this closing brace.
 */
void ASFormatter::formatClosingBrace(BraceType braceType)
{
	assert(!isBraceType(braceType, ARRAY_TYPE));
	assert(currentChar == '}');

	// parenStack must contain one entry
	if (parenStack->size() > 1)
		parenStack->pop_back();

	// mark state of immediately after empty block
	// this state will be used for locating braces that appear immediately AFTER an empty block (e.g. '{} \n}').
	if (previousCommandChar == '{')
		isImmediatelyPostEmptyBlock = true;

	if (attachClosingBraceMode)
	{
		// for now, namespaces and classes will be attached. Uncomment the lines below to break.
		if ((isEmptyLine(formattedLine)			// if a blank line precedes this
		        || isCharImmediatelyPostLineComment
		        || isCharImmediatelyPostComment
		        || (isImmediatelyPostPreprocessor && (int) currentLine.find_first_not_of(" \t") == charNum)
		    )
		        && (!isBraceType(braceType, SINGLE_LINE_TYPE) || isOkToBreakBlock(braceType)))
		{
			breakLine();
			appendCurrentChar();				// don't attach
		}
		else
		{
			if (previousNonWSChar != '{'
			        && (!isBraceType(braceType, SINGLE_LINE_TYPE)
			            || isOkToBreakBlock(braceType)))
				appendSpacePad();
			appendCurrentChar(false);			// attach
		}
	}
	else if (!isBraceType(braceType, EMPTY_BLOCK_TYPE)
	         && (isBraceType(braceType, BREAK_BLOCK_TYPE)
	             || isOkToBreakBlock(braceType)))
	{
		breakLine();
		appendCurrentChar();
	}
	else
	{
		appendCurrentChar();
	}

	// if a declaration follows a definition, space pad
	if (isLegalNameChar(peekNextChar()))
		appendSpaceAfter();

	if (shouldBreakBlocks
	        && currentHeader != nullptr
	        && !isHeaderInMultiStatementLine
	        && parenStack->back() == 0)
	{
		if (currentHeader == &ASResource::AS_CASE || currentHeader == &ASResource::AS_DEFAULT)
		{
			// do not yet insert a line if "break" statement is outside the braces
			std::string nextText = peekNextText(currentLine.substr(charNum + 1));
			if (!nextText.empty()
			        && nextText.substr(0, 5) != "break")
				isAppendPostBlockEmptyLineRequested = true;
		}
		else
		{
			// GH18
			// #569
			isAppendPostBlockEmptyLineRequested = !(shouldBreakBlocks && shouldAttachClosingWhile)
			                                      || currentHeader != &ASResource::AS_DO;
		}

	}
	// new option for break-blocks for classes and fcts? GL83
	else if (shouldBreakClosingHeaderBlocks)
	{
		isAppendPostBlockEmptyLineRequested = !currentHeader && shouldBreakBlocks;
	}
}

/**
 * format array braces as attached or broken
 * determine if the braces can have an inStatement indent
 * currentChar contains the brace
 * the braces will be appended to the current formattedLine or a new formattedLine as necessary
 * the calling function should have a continue statement after calling this method
 *
 * @param braceType            the type of brace to be formatted, must be an ARRAY_TYPE.
 * @param isOpeningArrayBrace  indicates if this is the opening brace for the array block.
 */
void ASFormatter::formatArrayBraces(BraceType braceType, bool isOpeningArrayBrace)
{
	assert(isBraceType(braceType, ARRAY_TYPE));
	assert(currentChar == '{' || currentChar == '}');

	if (currentChar == '{')
	{
		// is this the first opening brace in the array?
		if (isOpeningArrayBrace)
		{
			formatFirstOpenBrace(braceType);
		}
		else	     // not the first opening brace
		{
			formatOpenBrace();
		}
	}
	else if (currentChar == '}')
	{
		formatCloseBrace(braceType);
	}
}

/**
 * determine if a run-in can be attached.
 * if it can insert the indents in formattedLine and reset the current line break.
 */
void ASFormatter::formatRunIn()
{
	assert(braceFormatMode == RUN_IN_MODE || braceFormatMode == NONE_MODE);

	// keep one line blocks returns true without indenting the run-in
	if (formattingStyle != STYLE_PICO
	        && !isOkToBreakBlock(braceTypeStack->back()))
		return; // true;

	// make sure the line begins with a brace
	size_t lastText = formattedLine.find_last_not_of(" \t");
	if (lastText == std::string::npos || formattedLine[lastText] != '{')
		return; // false;

	// make sure the brace is broken
	if (formattedLine.find_first_not_of(" \t{") != std::string::npos)
		return; // false;

	if (isBraceType(braceTypeStack->back(), NAMESPACE_TYPE))
		return; // false;

	bool extraIndent = false;
	bool extraHalfIndent = false;
	isInLineBreak = true;

	// cannot attach a class modifier without indent-classes
	if (isCStyle()
	        && isCharPotentialHeader(currentLine, charNum)
	        && (isBraceType(braceTypeStack->back(), CLASS_TYPE)
	            || (isBraceType(braceTypeStack->back(), STRUCT_TYPE)
	                && isInIndentableStruct)))
	{
		if (findKeyword(currentLine, charNum, ASResource::AS_PUBLIC)
		        || findKeyword(currentLine, charNum, ASResource::AS_PRIVATE)
		        || findKeyword(currentLine, charNum, ASResource::AS_PROTECTED))
		{
			if (getModifierIndent())
				extraHalfIndent = true;
			else if (!getClassIndent())
				return; // false;
		}
		else if (getClassIndent())
			extraIndent = true;
	}

	// cannot attach a 'case' statement without indent-switches
	if (!getSwitchIndent()
	        && isCharPotentialHeader(currentLine, charNum)
	        && (findKeyword(currentLine, charNum, ASResource::AS_CASE)
	            || findKeyword(currentLine, charNum, ASResource::AS_DEFAULT)))
		return; // false;

	// extra indent for switch statements
	if (getSwitchIndent()
	        && !preBraceHeaderStack->empty()
	        && preBraceHeaderStack->back() == &ASResource::AS_SWITCH
	        && (isLegalNameChar(currentChar)
	            && !findKeyword(currentLine, charNum, ASResource::AS_CASE)))
		extraIndent = true;

	isInLineBreak = false;
	// remove for extra whitespace
	if (formattedLine.length() > lastText + 1
	        && formattedLine.find_first_not_of(" \t", lastText + 1) == std::string::npos)
		formattedLine.erase(lastText + 1);

	if (extraHalfIndent)
	{
		int indentLength_ = getIndentLength();
		runInIndentChars = indentLength_ / 2;
		formattedLine.append(runInIndentChars - 1, ' ');
	}
	else if (getForceTabIndentation() && getIndentLength() != getTabLength())
	{
		// insert the space indents
		std::string indent;
		int indentLength_ = getIndentLength();
		int tabLength_ = getTabLength();
		indent.append(indentLength_, ' ');
		if (extraIndent)
			indent.append(indentLength_, ' ');
		// replace spaces indents with tab indents
		size_t tabCount = indent.length() / tabLength_;		// truncate extra spaces
		indent.replace(0U, tabCount * tabLength_, tabCount, '\t');
		runInIndentChars = indentLength_;
		if (indent[0] == ' ')			// allow for brace
			indent.erase(0, 1);
		formattedLine.append(indent);
	}
	else if (getIndentString() == "\t")
	{
		appendChar('\t', false);
		runInIndentChars = 2;	// one for { and one for tab
		if (extraIndent)
		{
			appendChar('\t', false);
			runInIndentChars++;
		}
	}
	else // spaces
	{
		int indentLength_ = getIndentLength();
		formattedLine.append(indentLength_ - 1, ' ');
		runInIndentChars = indentLength_;
		if (extraIndent)
		{
			formattedLine.append(indentLength_, ' ');
			runInIndentChars += indentLength_;
		}
	}
	isInBraceRunIn = true;
}

/**
 * remove whitespace and add indentation for an array run-in.
 */
void ASFormatter::formatArrayRunIn()
{
	assert(isBraceType(braceTypeStack->back(), ARRAY_TYPE));

	// make sure the brace is broken
	if (formattedLine.find_first_not_of(" \t{") != std::string::npos)
		return;

	size_t lastText = formattedLine.find_last_not_of(" \t");
	if (lastText == std::string::npos || formattedLine[lastText] != '{')
		return;

	// check for extra whitespace
	if (formattedLine.length() > lastText + 1
	        && formattedLine.find_first_not_of(" \t", lastText + 1) == std::string::npos)
		formattedLine.erase(lastText + 1);

	if (getIndentString() == "\t")
	{
		appendChar('\t', false);
		runInIndentChars = 2;	// one for { and one for tab
	}
	else
	{
		int indent = getIndentLength();
		formattedLine.append(indent - 1, ' ');
		runInIndentChars = indent;
	}
	isInBraceRunIn = true;
	isInLineBreak = false;
}

/**
 * delete a braceTypeStack std::vector object
 * BraceTypeStack did not work with the DeleteContainer template
 */
void ASFormatter::deleteContainer(std::vector<BraceType>*& container)
{
	if (container != nullptr)
	{
		container->clear();
		delete (container);
		container = nullptr;
	}
}

/**
 * delete a std::vector object
 * T is the type of std::vector
 * used for all std::vectors except braceTypeStack
 */
template<typename T>
void ASFormatter::deleteContainer(T& container)
{
	if (container != nullptr)
	{
		container->clear();
		delete (container);
		container = nullptr;
	}
}

/**
 * initialize a braceType std::vector object
 * braceType did not work with the DeleteContainer template
 */
void ASFormatter::initContainer(std::vector<BraceType>*& container, std::vector<BraceType>* value)
{
	if (container != nullptr)
		deleteContainer(container);
	container = value;
}

/**
 * initialize a std::vector object
 * T is the type of std::vector
 * used for all std::vectors except braceTypeStack
 */
template<typename T>
void ASFormatter::initContainer(T& container, T value)
{
	// since the ASFormatter object is never deleted,
	// the existing std::vectors must be deleted before creating new ones
	if (container != nullptr)
		deleteContainer(container);
	container = value;
}

/**
 * convert a tab to spaces.
 * charNum points to the current character to convert to spaces.
 * tabIncrementIn is the increment that must be added for tab indent characters
 *     to get the correct column for the current tab.
 * replaces the tab in currentLine with the required number of spaces.
 * replaces the value of currentChar.
 */
void ASFormatter::convertTabToSpaces()
{
	assert(currentChar == '\t');

	// do NOT replace if in quotes
	if (isInQuote || isInQuoteContinuation)
		return;

	size_t tabSize = getTabLength();
	size_t numSpaces = tabSize - ((tabIncrementIn + charNum) % tabSize);
	currentLine.replace(charNum, 1, numSpaces, ' ');
	currentChar = currentLine[charNum];
}

/**
* is it ok to break this block?
*/
bool ASFormatter::isOkToBreakBlock(BraceType braceType) const
{
	// Actually, there should not be an ARRAY_TYPE brace here.
	// But this will avoid breaking a one line block when there is.
	// Otherwise they will be formatted differently on consecutive runs.
	if (isBraceType(braceType, ARRAY_TYPE)
	        && isBraceType(braceType, SINGLE_LINE_TYPE))
		return false;
	if (isBraceType(braceType, COMMAND_TYPE)
	        && isBraceType(braceType, EMPTY_BLOCK_TYPE))
		return false;
	if (!isBraceType(braceType, SINGLE_LINE_TYPE)
	        || isBraceType(braceType, BREAK_BLOCK_TYPE)
	        || shouldBreakOneLineBlocks)
		return true;
	return false;
}

/**
* check if a sharp header is a paren or non-paren header
*/
bool ASFormatter::isSharpStyleWithParen(const std::string* header) const
{
	return (isSharpStyle() && peekNextChar() == '('
	        && (header == &ASResource::AS_CATCH
	            || header == &ASResource::AS_DELEGATE));
}

/**
 * Check for a following header when a comment is reached.
 * firstLine must contain the start of the comment.
 * return value is a pointer to the header or nullptr.
 */
const std::string* ASFormatter::checkForHeaderFollowingComment(std::string_view firstLine) const
{
	assert(isInComment || isInLineComment);
	assert(shouldBreakElseIfs || shouldBreakBlocks || isInSwitchStatement());
	// look ahead to find the next non-comment text
	bool endOnEmptyLine = (currentHeader == nullptr);
	if (isInSwitchStatement())
		endOnEmptyLine = false;
	std::string nextText = peekNextText(firstLine, endOnEmptyLine);

	if (nextText.empty() || !isCharPotentialHeader(nextText, 0))
		return nullptr;

	return ASBase::findHeader(nextText, 0, headers);
}

/**
 * process preprocessor statements.
 * charNum should be the index of the #.
 *
 * delete braceTypeStack entries added by #if if a #else is found.
 * prevents double entries in the braceTypeStack.
 */
void ASFormatter::processPreprocessor()
{
	assert(currentChar == '#');

	const size_t preproc = currentLine.find_first_not_of(" \t", charNum + 1);
	if (preproc == std::string::npos)
		return;

	if (currentLine.compare(preproc, 2, "if") == 0)
	{
		preprocBraceTypeStackSize = braceTypeStack->size();
	}
	else if (currentLine.compare(preproc, 4, "else") == 0)
	{
		// delete stack entries added in #if
		// should be replaced by #else
		if (preprocBraceTypeStackSize > 0)
		{
			int addedPreproc = braceTypeStack->size() - preprocBraceTypeStackSize;
			for (int i = 0; i < addedPreproc; i++)
				braceTypeStack->pop_back();
		}
	}
	else if (currentLine.compare(preproc, 6, "define") == 0)
		isInPreprocessorDefineDef = true;


	//https://sourceforge.net/p/astyle/tickets/117/
	const size_t preprocPos = currentLine.find_first_not_of(" \t", charNum + 1);

	if (includeDirectivePaddingMode != INCLUDE_PAD_NO_CHANGE
	        && currentLine.compare(preprocPos, 7, "include") == 0)
	{
		size_t firstChar = currentLine.find_first_not_of(" \t", preprocPos + 7);
		if (firstChar != std::string::npos && (currentLine[firstChar] == '<' || currentLine[firstChar] == '"'))
		{
			currentLine.erase (preprocPos + 7, firstChar - (preprocPos + 7));
		}

		if (includeDirectivePaddingMode == INCLUDE_PAD_AFTER &&
		        (currentLine[preprocPos + 7] == '<' || currentLine[preprocPos + 7] == '"' || std::isalpha(currentLine[preprocPos + 7]))
		   )
		{
			currentLine.insert(preprocPos + 7, 1, ' ');
		}
	}

	// if top level it is potentially indentable
	if (shouldIndentPreprocBlock
	        && (isBraceType(braceTypeStack->back(), NULL_TYPE) || isBraceType(braceTypeStack->back(), NAMESPACE_TYPE))
	        && !foundClassHeader
	        && !isInClassInitializer
	        && sourceIterator->tellg() > preprocBlockEnd)
	{
		std::string preprocessor = ASBeautifier::extractPreprocessorStatement(currentLine);

		// indent the #if preprocessor blocks
		if (preprocessor.length() >= 2 && preprocessor.substr(0, 2) == "if") // #if, #ifdef, #ifndef
		{
			if (isImmediatelyPostPreprocessor)
				breakLine();
			isIndentablePreprocessorBlck = isIndentablePreprocessorBlock(currentLine, charNum);
			isIndentablePreprocessor = isIndentablePreprocessorBlck;
		}
	}

	if (isIndentablePreprocessorBlck
	        && charNum < (int) currentLine.length() - 1
	        && std::isblank(currentLine[charNum + 1]))
	{
		size_t nextText = currentLine.find_first_not_of(" \t", charNum + 1);
		if (nextText != std::string::npos)
		{
			currentLine.erase(charNum + 1, nextText - charNum - 1);
		}
	}

	if (isIndentablePreprocessorBlck
	        && sourceIterator->tellg() >= preprocBlockEnd)
		isIndentablePreprocessorBlck = false;
	//  need to fall thru here to reset the variables
}

/**
 * determine if the next line starts a comment
 * and a header follows the comment or comments.
 */
bool ASFormatter::commentAndHeaderFollows()
{
	// called ONLY IF shouldDeleteEmptyLines and shouldBreakBlocks are TRUE.
	assert(shouldDeleteEmptyLines && shouldBreakBlocks);

	// is the next line a comment
	auto stream = std::make_shared<ASPeekStream>(sourceIterator);
	if (!stream->hasMoreLines())
		return false;
	std::string nextLine_ = stream->peekNextLine();
	size_t firstChar = nextLine_.find_first_not_of(" \t");
	if (firstChar == std::string::npos
	        || !(nextLine_.compare(firstChar, 2, "//") == 0
	             || nextLine_.compare(firstChar, 2, "/*") == 0))
		return false;

	// find the next non-comment text, and reset
	std::string nextText = peekNextText(nextLine_, false, stream);
	if (nextText.empty() || !isCharPotentialHeader(nextText, 0))
		return false;

	const std::string* newHeader = ASBase::findHeader(nextText, 0, headers);

	if (newHeader == nullptr)
		return false;

	// if a closing header, reset break unless break is requested
	if (isClosingHeader(newHeader) && !shouldBreakClosingHeaderBlocks)
	{
		isAppendPostBlockEmptyLineRequested = false;
		return false;
	}

	return true;
}

/**
 * determine if a brace should be attached or broken
 * uses braces in the braceTypeStack
 * the last brace in the braceTypeStack is the one being formatted
 * returns true if the brace should be broken
 */
bool ASFormatter::isCurrentBraceBroken() const
{
	assert(braceTypeStack->size() > 1);

	bool breakBrace = false;
	size_t stackEnd = braceTypeStack->size() - 1;

	// check brace modifiers
	if (shouldAttachExternC
	        && isBraceType((*braceTypeStack)[stackEnd], EXTERN_TYPE))
	{
		return false;
	}
	if (shouldAttachNamespace
	        && isBraceType((*braceTypeStack)[stackEnd], NAMESPACE_TYPE))
	{
		return false;
	}
	if (shouldAttachClass
	        && (isBraceType((*braceTypeStack)[stackEnd], CLASS_TYPE)
	            || isBraceType((*braceTypeStack)[stackEnd], INTERFACE_TYPE)))
	{
		return false;
	}
	if (shouldAttachInline
	        && isCStyle()			// for C++ only
	        && braceFormatMode != RUN_IN_MODE
	        && !(currentLineBeginsWithBrace && peekNextChar() == '/')
	        && isBraceType((*braceTypeStack)[stackEnd], COMMAND_TYPE))
	{
		size_t i;
		for (i = 1; i < braceTypeStack->size(); i++)
			if (isBraceType((*braceTypeStack)[i], CLASS_TYPE)
			        || isBraceType((*braceTypeStack)[i], STRUCT_TYPE))
				return false;
	}

	// check braces
	if (isBraceType((*braceTypeStack)[stackEnd], EXTERN_TYPE))
	{
		if (currentLineBeginsWithBrace || braceFormatMode == RUN_IN_MODE)
		{
			breakBrace = true;
		}
	}
	else if (braceFormatMode == NONE_MODE)
	{
		if (currentLineBeginsWithBrace
		        && currentLineFirstBraceNum == (size_t) charNum)
		{
			breakBrace = true;
		}
	}
	else if (braceFormatMode == BREAK_MODE || braceFormatMode == RUN_IN_MODE)
	{
		breakBrace = true;
	}
	else if (braceFormatMode == LINUX_MODE)
	{
		// break a namespace
		if (isBraceType((*braceTypeStack)[stackEnd], NAMESPACE_TYPE))
		{
			if (formattingStyle != STYLE_STROUSTRUP
			        && formattingStyle != STYLE_MOZILLA
			        && formattingStyle != STYLE_WEBKIT)
			{
				breakBrace = true;
			}
		}
		// break a class or interface
		else if (isBraceType((*braceTypeStack)[stackEnd], CLASS_TYPE)
		         || isBraceType((*braceTypeStack)[stackEnd], INTERFACE_TYPE))
		{
			if (formattingStyle != STYLE_STROUSTRUP
			        && formattingStyle != STYLE_WEBKIT)
			{
				breakBrace = true;
			}
		}
		// break a struct if mozilla - an enum is processed as an array brace
		else if (isBraceType((*braceTypeStack)[stackEnd], STRUCT_TYPE))
		{
			if (formattingStyle == STYLE_MOZILLA)
			{
				breakBrace = true;
			}
		}
		// break the first brace if a function
		else if (isBraceType((*braceTypeStack)[stackEnd], COMMAND_TYPE))
		{
			if (stackEnd == 1)
			{
				breakBrace = true;
			}
			else if (stackEnd > 1)
			{
				// break the first brace after these if a function
				if (isBraceType((*braceTypeStack)[stackEnd - 1], NAMESPACE_TYPE)
				        || isBraceType((*braceTypeStack)[stackEnd - 1], CLASS_TYPE)
				        || (isBraceType((*braceTypeStack)[stackEnd - 1], ARRAY_TYPE) && !lambdaIndicator)
				        || isBraceType((*braceTypeStack)[stackEnd - 1], STRUCT_TYPE)
				        || isBraceType((*braceTypeStack)[stackEnd - 1], EXTERN_TYPE)
				   )
				{
					breakBrace = true;
				}
			}
		}
	}

	//breakBrace = false;
	return breakBrace;
}

/**
 * format comment body
 * the calling function should have a continue statement after calling this method
 */
void ASFormatter::formatCommentBody()
{
	assert(isInComment);

	// append the comment
	while (charNum < (int) currentLine.length())
	{
		currentChar = currentLine[charNum];
		if (isSequenceReached(ASResource::AS_CLOSE_COMMENT) || (isGSCStyle() && isSequenceReached(ASResource::AS_GSC_CLOSE_COMMENT)) )
		{
			formatCommentCloser();
			break;
		}
		if (currentChar == '\t' && shouldConvertTabs)
			convertTabToSpaces();
		appendCurrentChar();
		++charNum;
	}
	if (shouldStripCommentPrefix)
		stripCommentPrefix();
}

/**
 * format a comment opener
 * the comment opener will be appended to the current formattedLine or a new formattedLine as necessary
 * the calling function should have a continue statement after calling this method
 */
void ASFormatter::formatCommentOpener()
{
	assert(isSequenceReached(ASResource::AS_OPEN_COMMENT) || isSequenceReached(ASResource::AS_GSC_OPEN_COMMENT));

	bool isCppComment = isSequenceReached(ASResource::AS_OPEN_COMMENT);

	isInComment = isInCommentStartLine = true;
	isImmediatelyPostLineComment = false;
	if (previousNonWSChar == '}')
		resetEndOfStatement();

	// Check for a following header.
	// For speed do not check multiple comment lines more than once.
	// For speed do not check shouldBreakBlocks if previous line is empty, a comment, or a '{'.
	const std::string* followingHeader = nullptr;
	if ((doesLineStartComment
	        && !isImmediatelyPostCommentOnly
	        && isBraceType(braceTypeStack->back(), COMMAND_TYPE))
	        && (shouldBreakElseIfs
	            || isInSwitchStatement()
	            || (shouldBreakBlocks
	                && !isImmediatelyPostEmptyLine
	                && previousCommandChar != '{')))
		followingHeader = checkForHeaderFollowingComment(currentLine.substr(charNum));

	if (spacePadNum != 0 && !isInLineBreak)
		adjustComments();
	formattedLineCommentNum = formattedLine.length();

	// must be done BEFORE appendSequence
	if (previousCommandChar == '{'
	        && !isImmediatelyPostComment
	        && !isImmediatelyPostLineComment)
	{
		if (isBraceType(braceTypeStack->back(), NAMESPACE_TYPE))
		{
			// namespace run-in is always broken.
			isInLineBreak = true;
		}
		else if (braceFormatMode == NONE_MODE)
		{
			// should a run-in statement be attached?
			if (currentLineBeginsWithBrace)
				formatRunIn();
		}
		else if (braceFormatMode == ATTACH_MODE)
		{
			// if the brace was not attached?
			if (!formattedLine.empty() && formattedLine[0] == '{'
			        && !isBraceType(braceTypeStack->back(), SINGLE_LINE_TYPE))
				isInLineBreak = true;
		}
		else if (braceFormatMode == RUN_IN_MODE)
		{
			// should a run-in statement be attached?
			if (!formattedLine.empty() && formattedLine[0] == '{')
				formatRunIn();
		}
	}
	else if (!doesLineStartComment)
		noTrimCommentContinuation = true;

	// ASBeautifier needs to know the following statements
	if (shouldBreakElseIfs && followingHeader == &ASResource::AS_ELSE)
		elseHeaderFollowsComments = true;
	if (followingHeader == &ASResource::AS_CASE || followingHeader == &ASResource::AS_DEFAULT)
		caseHeaderFollowsComments = true;

	// appendSequence will write the previous line
	appendSequence(isCppComment ? ASResource::AS_OPEN_COMMENT : ASResource::AS_GSC_OPEN_COMMENT);
	goForward(1);

	// must be done AFTER appendSequence

	// Break before the comment if a header follows the line comment.
	// But not break if previous line is empty, a comment, or a '{'.
	if (shouldBreakBlocks
	        && followingHeader != nullptr
	        && !isImmediatelyPostEmptyLine
	        && previousCommandChar != '{')
	{
		if (isClosingHeader(followingHeader))
		{
			if (!shouldBreakClosingHeaderBlocks)
				isPrependPostBlockEmptyLineRequested = false;
		}
		// if an opening header, break before the comment
		else
			isPrependPostBlockEmptyLineRequested = true;
	}

	if (previousCommandChar == '}')
		currentHeader = nullptr;
}

/**
 * format a comment closer
 * the comment closer will be appended to the current formattedLine
 */
void ASFormatter::formatCommentCloser()
{
	assert(isSequenceReached(ASResource::AS_CLOSE_COMMENT) || (isGSCStyle() && isSequenceReached(ASResource::AS_GSC_CLOSE_COMMENT)) );
	isInComment = false;
	noTrimCommentContinuation = false;
	isImmediatelyPostComment = true;
	appendSequence(isSequenceReached(ASResource::AS_CLOSE_COMMENT) ? ASResource::AS_CLOSE_COMMENT : ASResource::AS_GSC_CLOSE_COMMENT);
	goForward(1);
	if (doesLineStartComment
	        && (currentLine.find_first_not_of(" \t", charNum + 1) == std::string::npos))
		lineEndsInCommentOnly = true;
	if (peekNextChar() == '}'
	        && previousCommandChar != ';'
	        && !isBraceType(braceTypeStack->back(), ARRAY_TYPE)
	        && !isInPreprocessor
	        && isOkToBreakBlock(braceTypeStack->back()))
	{
		isInLineBreak = true;
		shouldBreakLineAtNextChar = true;
	}
}

/**
 * format a line comment body
 * the calling function should have a continue statement after calling this method
 */
void ASFormatter::formatLineCommentBody()
{
	assert(isInLineComment);

	// append the comment
	while (charNum < (int) currentLine.length())
//	        && !isLineReady	// commented out in release 2.04, unnecessary
	{
		currentChar = currentLine[charNum];
		if (currentChar == '\t' && shouldConvertTabs)
			convertTabToSpaces();
		appendCurrentChar();
		++charNum;
	}

	// explicitly break a line when a line comment's end is found.
	if (charNum == (int) currentLine.length())
	{
		isInLineBreak = true;
		isInLineComment = false;
		isImmediatelyPostLineComment = true;
		currentChar = 0;  //make sure it is a neutral char.
	}
}

/**
 * format a line comment opener
 * the line comment opener will be appended to the current formattedLine or a new formattedLine as necessary
 * the calling function should have a continue statement after calling this method
 */
void ASFormatter::formatLineCommentOpener()
{
	assert(isSequenceReached(ASResource::AS_OPEN_LINE_COMMENT));

	if ((int) currentLine.length() > charNum + 2
	        && currentLine[charNum + 2] == '\xf2')     // check for windows line marker
		isAppendPostBlockEmptyLineRequested = false;

	isInLineComment = true;
	isCharImmediatelyPostComment = false;
	if (previousNonWSChar == '}')
		resetEndOfStatement();

	// Check for a following header.
	// For speed do not check multiple comment lines more than once.
	// For speed do not check shouldBreakBlocks if previous line is empty, a comment, or a '{'.
	const std::string* followingHeader = nullptr;
	if ((lineIsLineCommentOnly
	        && !isImmediatelyPostCommentOnly
	        && isBraceType(braceTypeStack->back(), COMMAND_TYPE))
	        && (shouldBreakElseIfs
	            || isInSwitchStatement()
	            || (shouldBreakBlocks
	                && !isImmediatelyPostEmptyLine
	                && previousCommandChar != '{')))
		followingHeader = checkForHeaderFollowingComment(currentLine.substr(charNum));

	// do not indent if in column 1 or 2
	// or in a namespace before the opening brace
	if ((!shouldIndentCol1Comments && !lineCommentNoIndent)
	        || foundNamespaceHeader)
	{
		if (charNum == 0)
			lineCommentNoIndent = true;
		else if (charNum == 1 && currentLine[0] == ' ')
			lineCommentNoIndent = true;
	}
	// move comment if spaces were added or deleted
	if (!lineCommentNoIndent && spacePadNum != 0 && !isInLineBreak)
		adjustComments();
	formattedLineCommentNum = formattedLine.length();

	// must be done BEFORE appendSequence
	// check for run-in statement
	if (previousCommandChar == '{'
	        && !isImmediatelyPostComment
	        && !isImmediatelyPostLineComment)
	{
		if (braceFormatMode == NONE_MODE)
		{
			if (currentLineBeginsWithBrace)
				formatRunIn();
		}
		else if (braceFormatMode == RUN_IN_MODE)
		{
			if (!lineCommentNoIndent)
				formatRunIn();
			else
				isInLineBreak = true;
		}
		else if (braceFormatMode == BREAK_MODE)
		{
			if (!formattedLine.empty() && formattedLine[0] == '{')
				isInLineBreak = true;
		}
		else
		{
			if (currentLineBeginsWithBrace)
				isInLineBreak = true;
		}
	}

	// ASBeautifier needs to know the following statements
	if (shouldBreakElseIfs && followingHeader == &ASResource::AS_ELSE)
		elseHeaderFollowsComments = true;
	if (followingHeader == &ASResource::AS_CASE || followingHeader == &ASResource::AS_DEFAULT)
		caseHeaderFollowsComments = true;

	// appendSequence will write the previous line
	appendSequence(ASResource::AS_OPEN_LINE_COMMENT);
	goForward(1);

	// must be done AFTER appendSequence

	// Break before the comment if a header follows the line comment.
	// But do not break if previous line is empty, a comment, or a '{'.
	if (shouldBreakBlocks
	        && followingHeader != nullptr
	        && !isImmediatelyPostEmptyLine
	        && previousCommandChar != '{')
	{
		if (isClosingHeader(followingHeader))
		{
			if (!shouldBreakClosingHeaderBlocks)
				isPrependPostBlockEmptyLineRequested = false;
		}
		// if an opening header, break before the comment
		else
			isPrependPostBlockEmptyLineRequested = true;
	}

	if (previousCommandChar == '}')
		currentHeader = nullptr;

	// if tabbed input don't convert the immediately following tabs to spaces
	if (getIndentString() == "\t" && lineCommentNoIndent)
	{
		while (charNum + 1 < (int) currentLine.length()
		        && currentLine[charNum + 1] == '\t')
		{
			currentChar = currentLine[++charNum];
			appendCurrentChar();
		}
	}

	// explicitly break a line when a line comment's end is found.
	if (charNum + 1 == (int) currentLine.length())
	{
		isInLineBreak = true;
		isInLineComment = false;
		isImmediatelyPostLineComment = true;
		currentChar = 0;  //make sure it is a neutral char.
	}
}

/**
 * format quote body
 * the calling function should have a continue statement after calling this method
 */
void ASFormatter::formatQuoteBody()
{
	assert(isInQuote);

	int _braceCount = 0;

	if (checkInterpolation && currentChar == '{')
	{
		++_braceCount;
	}

	if (isSpecialChar)
	{
		isSpecialChar = false;
	}
	else if (currentChar == '\\' && !isInVerbatimQuote)
	{
		if (peekNextChar() == ' ')              // is this '\' at end of line
			haveLineContinuationChar = true;
		else
			isSpecialChar = true;
	}
	else if (isInVerbatimQuote && currentChar == '"' )
	{
		if (isCStyle())
		{
			std::string delim = ')' + verbatimDelimiter;
			int delimStart = charNum - delim.length();
			if (delimStart > 0 && currentLine.substr(delimStart, delim.length()) == delim)
			{
				isInQuote = false;
				isInVerbatimQuote = false;
				checkInterpolation = false;
			}
		}
		else if (isSharpStyle() )       // GH16
		{
			if ((int) currentLine.length() > charNum + 1
			        && currentLine[charNum + 1] == '"')			// check consecutive quotes
			{
				appendSequence("\"\"");
				goForward(1);
				return;
			}

			//if ( charNum>0 && currentLine[charNum - 1] != '\\')
			isInQuote = false;

			if (checkInterpolation)
				isInVerbatimQuote = false;

			checkInterpolation = false;
		}
	}
	else if (quoteChar == currentChar)
	{
		////do not quit if we have CS std::string with interpolation
		isInQuote = false;
	}

	appendCurrentChar();

	// append the text to the ending quoteChar or an escape sequence
	// tabs in quotes are NOT changed by convert-tabs
	if (isInQuote && currentChar != '\\')
	{
		while (charNum + 1 < (int) currentLine.length()
		        && ( currentLine[charNum + 1] != quoteChar || _braceCount > 0 )
		        && currentLine[charNum + 1] != '\\')
		{
			currentChar = currentLine[++charNum];

			if (checkInterpolation)
			{
				if (currentChar == '{')
					++_braceCount;

				if (currentChar == '}')
					--_braceCount;
			}
			appendCurrentChar();
		}
	}
	if (charNum + 1 >= (int) currentLine.length()
	        && currentChar != '\\'
	        && !isInVerbatimQuote)
	{
		isInQuote = false;				// missing closing quote
	}

}

/**
 * format a quote opener
 * the quote opener will be appended to the current formattedLine or a new formattedLine as necessary
 * the calling function should have a continue statement after calling this method
 */
void ASFormatter::formatQuoteOpener()
{
	assert(currentChar == '"'
	       || (currentChar == '\'' && !isDigitSeparator(currentLine, charNum)));

	isInQuote = true;
	quoteChar = currentChar;

	char prevPrevCh = charNum > 2 ? currentLine[charNum - 2] : ' '; // GL39
	if (isCStyle() && previousChar == 'R' && !isalpha(prevPrevCh))
	{
		int parenPos = currentLine.find('(', charNum);
		if (parenPos != -1)
		{
			isInVerbatimQuote = true;
			verbatimDelimiter = currentLine.substr(charNum + 1, parenPos - charNum - 1);
		}
	}
	else if (isSharpStyle() && (previousChar == '@' ))
	{
		isInVerbatimQuote = true;
		checkInterpolation = true;
	}


	// a quote following a brace is an array
	if (previousCommandChar == '{'
	        && !isImmediatelyPostComment
	        && !isImmediatelyPostLineComment
	        && isNonInStatementArray
	        && !isBraceType(braceTypeStack->back(), SINGLE_LINE_TYPE)
	        && !std::isblank(peekNextChar()))
	{
		if (braceFormatMode == NONE_MODE)
		{
			if (currentLineBeginsWithBrace)
				formatRunIn();
		}
		else if (braceFormatMode == RUN_IN_MODE)
		{
			formatRunIn();
		}
		else if (braceFormatMode == BREAK_MODE)
		{
			if (!formattedLine.empty() && formattedLine[0] == '{')
				isInLineBreak = true;
		}
		else
		{
			if (currentLineBeginsWithBrace)
				isInLineBreak = true;
		}
	}
	previousCommandChar = ' ';
	appendCurrentChar();
}

/**
 * get the next line comment adjustment that results from breaking a closing brace.
 * the brace must be on the same line as the closing header.
 * i.e "} else" changed to "} <NL> else".
 */
int ASFormatter::getNextLineCommentAdjustment()
{
	assert(foundClosingHeader && previousNonWSChar == '}');
	if (charNum < 1)			// "else" is in column 1
		return 0;
	size_t lastBrace = currentLine.rfind('}', charNum - 1);
	if (lastBrace != std::string::npos)
		return (lastBrace - charNum);	// return a negative number
	return 0;
}

// for console build only
LineEndFormat ASFormatter::getLineEndFormat() const
{
	return lineEnd;
}

/**
 * get the current line comment adjustment that results from attaching
 * a closing header to a closing brace.
 * the brace must be on the line previous to the closing header.
 * the adjustment is 2 chars, one for the brace and one for the space.
 * i.e "} <NL> else" changed to "} else".
 */
int ASFormatter::getCurrentLineCommentAdjustment()
{
	assert(foundClosingHeader && previousNonWSChar == '}');
	if (charNum < 1)
		return 2;
	size_t lastBrace = currentLine.rfind('}', charNum - 1);
	if (lastBrace == std::string::npos)
		return 2;
	return 0;
}

/**
 * get the previous word on a line
 * the argument 'currPos' must point to the current position.
 *
 * @return is the previous word or an empty std::string if none found.
 */
std::string ASFormatter::getPreviousWord(const std::string& line, int currPos, bool allowDots) const
{
	// get the last legal word (may be a number)
	if (currPos == 0)
		return std::string();

	size_t end = line.find_last_not_of(" \t", currPos - 1);
	if (end == std::string::npos || !isLegalNameChar(line[end]))
		return std::string();

	int start;          // start of the previous word
	for (start = end; start > -1; start--)
	{
		if (!isLegalNameChar(line[start]) || (!allowDots && line[start] == '.') )
			break;
	}
	start++;

	return (line.substr(start, end - start + 1));
}

/**
 * check if a line break is needed when a closing brace
 * is followed by a closing header.
 * the break depends on the braceFormatMode and other factors.
 */
void ASFormatter::isLineBreakBeforeClosingHeader()
{
	assert(foundClosingHeader && previousNonWSChar == '}');

	if (currentHeader == &ASResource::AS_WHILE && shouldAttachClosingWhile)
	{
		appendClosingHeader();
		return;
	}

	if (braceFormatMode == BREAK_MODE
	        || braceFormatMode == RUN_IN_MODE
	        || attachClosingBraceMode)
	{
		isInLineBreak = true;
	}
	else if (braceFormatMode == NONE_MODE)
	{
		if (shouldBreakClosingHeaderBraces
		        || getBraceIndent() || getBlockIndent())
		{
			isInLineBreak = true;
		}
		else
		{
			appendSpacePad();
			// is closing brace broken?
			size_t i = currentLine.find_first_not_of(" \t");
			if (i != std::string::npos && currentLine[i] == '}')
				isInLineBreak = false;

			if (shouldBreakBlocks)
				isAppendPostBlockEmptyLineRequested = false;
		}
	}
	// braceFormatMode == ATTACH_MODE, LINUX_MODE
	else
	{
		if (shouldBreakClosingHeaderBraces
		        || getBraceIndent() || getBlockIndent())
		{
			isInLineBreak = true;
		}
		else
		{
			appendClosingHeader();
			if (shouldBreakBlocks)
				isAppendPostBlockEmptyLineRequested = false;
		}
	}
}

/**
 * Append a closing header to the previous closing brace, if possible
 */
void ASFormatter::appendClosingHeader()
{
	// if a blank line does not precede this
	// or last line is not a one line block, attach header
	bool previousLineIsEmpty = isEmptyLine(formattedLine);
	int previousLineIsOneLineBlock = 0;
	size_t firstBrace = findNextChar(formattedLine, '{');
	if (firstBrace != std::string::npos)
		previousLineIsOneLineBlock = isOneLineBlockReached(formattedLine, firstBrace);
	if (!previousLineIsEmpty
	        && previousLineIsOneLineBlock == 0)
	{
		isInLineBreak = false;
		appendSpacePad();
		spacePadNum = 0;	// don't count as comment padding
	}
}

/**
 * Add braces to a single line statement following a header.
 * braces are not added if the proper conditions are not met.
 * braces are added to the currentLine.
 */
bool ASFormatter::addBracesToStatement()
{
	assert(isImmediatelyPostHeader);

	if (currentHeader != &ASResource::AS_IF
	        && currentHeader != &ASResource::AS_ELSE
	        && currentHeader != &ASResource::AS_FOR
	        && currentHeader != &ASResource::AS_WHILE
	        && currentHeader != &ASResource::AS_DO
	        && currentHeader != &ASResource::AS_FOREACH
	        && currentHeader != &ASResource::AS_QFOREACH
	        && currentHeader != &ASResource::AS_QFOREVER
	        && currentHeader != &ASResource::AS_FOREVER)
		return false;

	if (currentHeader == &ASResource::AS_WHILE && foundClosingHeader)	// do-while
		return false;

	// do not brace an empty statement
	if (currentChar == ';')
		return false;


	// old behavior
	if (shouldAddBraces)
	{

		// do not add if a header follows
		if (isCharPotentialHeader(currentLine, charNum))
			if (findHeader(headers) != nullptr)
				return false;

		// find the next semi-colon
		size_t nextSemiColon = charNum;
		if (currentChar != ';')
			nextSemiColon = findNextChar(currentLine, ';', charNum + 1);
		if (nextSemiColon == std::string::npos)
			return false;

		// add closing brace before changing the line length
		if (nextSemiColon == currentLine.length() - 1)
			currentLine.append(" }");
		else
			currentLine.insert(nextSemiColon + 1, " }");
	}

	// add opening brace
	currentLine.insert(charNum, "{ ");
	assert(computeChecksumIn("{}"));
	currentChar = '{';
	if ((int) currentLine.find_first_not_of(" \t") == charNum)
		currentLineBeginsWithBrace = true;
	// remove extra spaces
	if (!shouldAddOneLineBraces)
	{
		size_t lastText = formattedLine.find_last_not_of(" \t");
		if ((formattedLine.length() - 1) - lastText > 1)
			formattedLine.erase(lastText + 1);
	}
	return true;
}

/**
 * Remove braces from a single line statement following a header.
 * braces are not removed if the proper conditions are not met.
 * The first brace is replaced by a space.
 */
bool ASFormatter::removeBracesFromStatement()
{
	assert(isImmediatelyPostHeader);
	assert(currentChar == '{');

	if (currentHeader != &ASResource::AS_IF
	        && currentHeader != &ASResource::AS_ELSE
	        && currentHeader != &ASResource::AS_FOR
	        && currentHeader != &ASResource::AS_WHILE
	        && currentHeader != &ASResource::AS_FOREACH)
		return false;

	if (currentHeader == &ASResource::AS_WHILE && foundClosingHeader)	// do-while
		return false;

	bool isFirstLine = true;
	std::string nextLine_;
	// leave nextLine_ empty if end of line comment follows
	if (!isBeforeAnyLineEndComment(charNum) || currentLineBeginsWithBrace)
		nextLine_ = currentLine.substr(charNum + 1);
	size_t nextChar = 0;

	// find the first non-blank text
	ASPeekStream stream(sourceIterator);
	while (stream.hasMoreLines() || isFirstLine)
	{
		if (isFirstLine)
			isFirstLine = false;
		else
		{
			nextLine_ = stream.peekNextLine();
			nextChar = 0;
		}

		nextChar = nextLine_.find_first_not_of(" \t", nextChar);
		if (nextChar != std::string::npos)
			break;
	}
	if (!stream.hasMoreLines())
		return false;

	// don't remove if comments or a header follow the brace
	if ((nextLine_.compare(nextChar, 2, "/*") == 0)
	        || (nextLine_.compare(nextChar, 2, "//") == 0)
	        || (isCharPotentialHeader(nextLine_, nextChar)
	            && ASBase::findHeader(nextLine_, nextChar, headers) != nullptr))
		return false;

	// find the next semi-colon
	size_t nextSemiColon = nextChar;
	if (nextLine_[nextChar] != ';')
		nextSemiColon = findNextChar(nextLine_, ';', nextChar + 1);
	if (nextSemiColon == std::string::npos)
		return false;

	// find the closing brace
	isFirstLine = true;
	nextChar = nextSemiColon + 1;
	while (stream.hasMoreLines() || isFirstLine)
	{
		if (isFirstLine)
			isFirstLine = false;
		else
		{
			nextLine_ = stream.peekNextLine();
			nextChar = 0;
		}
		nextChar = nextLine_.find_first_not_of(" \t", nextChar);
		if (nextChar != std::string::npos)
			break;
	}
	if (nextLine_.empty() || nextLine_[nextChar] != '}')
		return false;

	// remove opening brace
	currentLine[charNum] = currentChar = ' ';
	assert(adjustChecksumIn(-'{'));
	return true;
}

/**
 * Find the next character that is not in quotes or a comment.
 *
 * @param line         the line to be searched.
 * @param searchChar   the char to find.
 * @param searchStart  the start position on the line (default is 0).
 * @return the position on the line or std::string::npos if not found.
 */
size_t ASFormatter::findNextChar(std::string_view line, char searchChar, int searchStart /*0*/) const
{
	// find the next searchChar
	size_t i;
	for (i = searchStart; i < line.length(); i++)
	{
		if (line.compare(i, 2, "//") == 0)
			return std::string::npos;
		if (line.compare(i, 2, "/*") == 0)
		{
			size_t endComment = line.find("*/", i + 2);
			if (endComment == std::string::npos)
				return std::string::npos;
			i = endComment + 2;
			if (i >= line.length())
				return std::string::npos;
		}
		if (line[i] == '"'
		        || (line[i] == '\'' && !isDigitSeparator(line, i)))
		{
			char quote = line[i];
			while (i < line.length())
			{
				size_t endQuote = line.find(quote, i + 1);
				if (endQuote == std::string::npos)
					return std::string::npos;
				i = endQuote;
				if (line[endQuote - 1] != '\\')	// check for '\"'
					break;
				if (line[endQuote - 2] == '\\')	// check for '\\'
					break;
			}
		}

		if (line[i] == searchChar)
			break;

		// for now don't process C# 'delegate' braces
		// do this last in case the search char is a '{'
		if (line[i] == '{')
			return std::string::npos;
	}
	if (i >= line.length())	// didn't find searchChar
		return std::string::npos;

	return i;
}

/**
 * Find split point for break/attach return type.
 */
void ASFormatter::findReturnTypeSplitPoint(const std::string& firstLine)
{
	assert((isBraceType(braceTypeStack->back(), NULL_TYPE)
	        || isBraceType(braceTypeStack->back(), DEFINITION_TYPE)));
	assert(shouldBreakReturnType || shouldBreakReturnTypeDecl
	       || shouldAttachReturnType || shouldAttachReturnTypeDecl);

	bool isFirstLine     = true;
	bool isInComment_    = false;
	bool isInQuote_      = false;
	bool foundSplitPoint = false;
	bool isAlreadyBroken = false;
	char quoteChar_      = ' ';
	char currNonWSChar   = ' ';
	char prevNonWSChar   = ' ';
	size_t parenCount    = 0;
	size_t squareCount   = 0;
	size_t angleCount    = 0;
	size_t breakLineNum  = 0;
	size_t breakCharNum  = std::string::npos;
	std::string line          = firstLine;

	// Process the lines until a ';' or '{'.
	ASPeekStream stream(sourceIterator);
	while (stream.hasMoreLines() || isFirstLine)
	{
		if (isFirstLine)
			isFirstLine = false;
		else
		{
			if (isInQuote_)
				return;
			line = stream.peekNextLine();
			if (!foundSplitPoint)
				++breakLineNum;
		}
		size_t firstCharNum = line.find_first_not_of(" \t");
		if (firstCharNum == std::string::npos)
			continue;
		if (line[firstCharNum] == '#')
		{
			// don't attach to a preprocessor
			if (shouldAttachReturnType || shouldAttachReturnTypeDecl)
				return;
			continue;
		}
		// parse the line
		for (size_t i = 0; i < line.length(); i++)
		{
			if (!std::isblank(line[i]))
			{
				prevNonWSChar = currNonWSChar;
				currNonWSChar = line[i];
			}
			else if (line[i] == '\t' && shouldConvertTabs)
			{
				size_t tabSize = getTabLength();
				size_t numSpaces = tabSize - ((tabIncrementIn + i) % tabSize);
				line.replace(i, 1, numSpaces, ' ');
				currentChar = line[i];
			}
			if (line.compare(i, 2, "/*") == 0)
				isInComment_ = true;
			if (isInComment_)
			{
				if (line.compare(i, 2, "*/") == 0)
				{
					isInComment_ = false;
					++i;
				}
				continue;
			}
			if (line[i] == '\\')
			{
				++i;
				continue;
			}

			if (isInQuote_)
			{
				if (line[i] == quoteChar_)
					isInQuote_ = false;
				continue;
			}

			if (line[i] == '"'
			        || (line[i] == '\'' && !isDigitSeparator(line, i)))
			{
				isInQuote_ = true;
				quoteChar_ = line[i];
				continue;
			}
			if (line.compare(i, 2, "//") == 0)
			{
				i = line.length();
				continue;
			}

			// https://sourceforge.net/p/astyle/bugs/504/
			if (line[line.length() - 1] == ':')
			{
				i = line.length();
				foundSplitPoint = true;
				continue;
			}

			// not in quote or comment
			if (!foundSplitPoint)
			{
				if (line[i] == '<')
				{
					++angleCount;
					continue;
				}
				if (line[i] == '>')
				{
					if (angleCount)
						--angleCount;
					if (!angleCount)
					{
						size_t nextCharNum = line.find_first_not_of(" \t*&", i + 1);
						if (nextCharNum == std::string::npos)
						{
							breakCharNum  = std::string::npos;
							continue;
						}
						if (line[nextCharNum] != ':')		// scope operator
							breakCharNum  = nextCharNum;
					}
					continue;
				}
				if (angleCount)
					continue;
				if (line[i] == '[')
				{
					++squareCount;
					continue;
				}
				if (line[i] == ']')
				{
					if (squareCount)
						--squareCount;
					continue;
				}
				// an assignment before the parens is not a function
				if (line[i] == '=')
					return;
				if (std::isblank(line[i]) || line[i] == '*' || line[i] == '&')
				{
					size_t nextNum = line.find_first_not_of(" \t", i + 1);
					if (nextNum == std::string::npos)
						breakCharNum = std::string::npos;
					else
					{
						if (line.length() > nextNum + 1
						        && line[nextNum] == ':' && line[nextNum + 1] == ':')
							i = --nextNum;
						else if (line[nextNum] != '(')
							breakCharNum = std::string::npos;
					}
					continue;
				}
				if ((isLegalNameChar(line[i]) || line[i] == '~')
				        && breakCharNum == std::string::npos)
				{
					breakCharNum = i;
					if (isLegalNameChar(line[i])
					        && findKeyword(line, i, ASResource::AS_OPERATOR))
					{
						if (breakCharNum == firstCharNum)
							isAlreadyBroken = true;
						foundSplitPoint = true;
						// find the operator, may be parens
						size_t parenNum =
						    line.find_first_not_of(" \t", i + ASResource::AS_OPERATOR.length());
						if (parenNum == std::string::npos)
							return;
						// find paren after the operator
						parenNum = line.find('(', parenNum + 1);
						if (parenNum == std::string::npos)
							return;
						i = --parenNum;
					}
					continue;
				}
				if (line[i] == ':'
				        && line.length() > i + 1
				        && line[i + 1] == ':')
				{
					size_t nextCharNum = line.find_first_not_of(" \t:", i + 1);
					if (nextCharNum == std::string::npos)
						return;

					if (isLegalNameChar(line[nextCharNum])
					        && findKeyword(line, nextCharNum, ASResource::AS_OPERATOR))
					{
						i = nextCharNum;
						if (breakCharNum == firstCharNum)
							isAlreadyBroken = true;
						foundSplitPoint = true;
						// find the operator, may be parens
						size_t parenNum =
						    line.find_first_not_of(" \t", i + ASResource::AS_OPERATOR.length());
						if (parenNum == std::string::npos)
							return;
						// find paren after the operator
						parenNum = line.find('(', parenNum + 1);
						if (parenNum == std::string::npos)
							return;
						i = --parenNum;
					}
					else
						i = --nextCharNum;
					continue;
				}
				if (line[i] == '(' && !squareCount)
				{

					// is line is already broken?
					if (breakCharNum == firstCharNum && breakLineNum > 0)
						isAlreadyBroken = true;
					++parenCount;
					foundSplitPoint = true;
					continue;
				}
			}
			// end !foundSplitPoint
			if (line[i] == '(')
			{
				// consecutive ')(' parens is probably a function pointer
				if (prevNonWSChar == ')' && !parenCount)
					return;
				++parenCount;
				continue;
			}
			if (line[i] == ')')
			{
				if (parenCount)
					--parenCount;
				continue;
			}
			if (line[i] == '{')
			{
				if (shouldBreakReturnType && foundSplitPoint && !isAlreadyBroken)
				{
					methodBreakCharNum = breakCharNum;
					methodBreakLineNum = breakLineNum;
				}

				if (shouldAttachReturnType && foundSplitPoint && isAlreadyBroken)
				{
					//https://sourceforge.net/p/astyle/bugs/545/
					if ((maxCodeLength != std::string::npos && previousReadyFormattedLineLength < maxCodeLength) || maxCodeLength == std::string::npos)
					{
						methodAttachCharNum = breakCharNum;
						methodAttachLineNum = breakLineNum;
					}
				}
				return;
			}
			if (line[i] == ';')
			{
				if (shouldBreakReturnTypeDecl && foundSplitPoint && !isAlreadyBroken)
				{
					methodBreakCharNum = breakCharNum;
					methodBreakLineNum = breakLineNum;
				}
				if (shouldAttachReturnTypeDecl && foundSplitPoint && isAlreadyBroken)
				{
					methodAttachCharNum = breakCharNum;
					methodAttachLineNum = breakLineNum;
				}
				return;
			}
			if (line[i] == '}')
				return;
		}   // end of for loop
		if (!foundSplitPoint)
			breakCharNum = std::string::npos;
	}   // end of while loop
}

/**
 * Look ahead in the file to see if a struct has access modifiers.
 *
 * @param firstLine     a reference to the line to indent.
 * @param index         the current line index.
 * @return              true if the struct has access modifiers.
 */
bool ASFormatter::isStructAccessModified(const std::string& firstLine, size_t index) const
{
	assert(firstLine[index] == '{');
	assert(isCStyle());

	bool isFirstLine = true;
	size_t braceCount = 1;
	std::string nextLine_ = firstLine.substr(index + 1);
	ASPeekStream stream(sourceIterator);

	// find the first non-blank text, bypassing all comments and quotes.
	bool isInComment_ = false;
	bool isInQuote_ = false;
	char quoteChar_ = ' ';
	while (stream.hasMoreLines() || isFirstLine)
	{
		if (isFirstLine)
			isFirstLine = false;
		else
			nextLine_ = stream.peekNextLine();
		// parse the line
		for (size_t i = 0; i < nextLine_.length(); i++)
		{
			if (std::isblank(nextLine_[i]))
				continue;
			if (nextLine_.compare(i, 2, "/*") == 0)
				isInComment_ = true;
			if (isInComment_)
			{
				if (nextLine_.compare(i, 2, "*/") == 0)
				{
					isInComment_ = false;
					++i;
				}
				continue;
			}
			if (nextLine_[i] == '\\')
			{
				++i;
				continue;
			}

			if (isInQuote_)
			{
				if (nextLine_[i] == quoteChar_)
					isInQuote_ = false;
				continue;
			}

			if (nextLine_[i] == '"'
			        || (nextLine_[i] == '\'' && !isDigitSeparator(nextLine_, i)))
			{
				isInQuote_ = true;
				quoteChar_ = nextLine_[i];
				continue;
			}
			if (nextLine_.compare(i, 2, "//") == 0)
			{
				i = nextLine_.length();
				continue;
			}
			// handle braces
			if (nextLine_[i] == '{')
				++braceCount;
			if (nextLine_[i] == '}')
				--braceCount;
			if (braceCount == 0)
				return false;
			// check for access modifiers
			if (isCharPotentialHeader(nextLine_, i))
			{
				if (findKeyword(nextLine_, i, ASResource::AS_PUBLIC)
				        || findKeyword(nextLine_, i, ASResource::AS_PRIVATE)
				        || findKeyword(nextLine_, i, ASResource::AS_PROTECTED))
					return true;
				std::string_view name = getCurrentWord(nextLine_, i);
				i += name.length() - 1;
			}
		}	// end of for loop
	}	// end of while loop

	return false;
}

/**
* Look ahead in the file to see if a preprocessor block is indentable.
*
* @param firstLine     a reference to the line to indent.
* @param index         the current line index.
* @return              true if the block is indentable.
*/
bool ASFormatter::isIndentablePreprocessorBlock(const std::string& firstLine, size_t index)
{
	assert(firstLine[index] == '#');

	bool isFirstLine = true;
	bool isInIndentableBlock = false;
	bool blockContainsBraces = false;
	bool blockContainsDefineContinuation = false;
	bool isInClassConstructor = false;
	bool isPotentialHeaderGuard = false;	// ifndef is first preproc statement
	bool isPotentialHeaderGuard2 = false;	// define is within the first preproc
	int  numBlockIndents = 0;
	int  lineParenCount = 0;
	std::string nextLine_ = firstLine.substr(index);
	auto stream = std::make_shared<ASPeekStream>(sourceIterator);

	// find end of the block, bypassing all comments and quotes.
	bool isInComment_ = false;
	bool isInQuote_ = false;
	char quoteChar_ = ' ';
	while (stream->hasMoreLines() || isFirstLine)
	{
		if (isFirstLine)
			isFirstLine = false;
		else
			nextLine_ = stream->peekNextLine();
		// parse the line
		for (size_t i = 0; i < nextLine_.length(); i++)
		{
			if (std::isblank(nextLine_[i]))
				continue;
			if (nextLine_.compare(i, 2, "/*") == 0)
				isInComment_ = true;
			if (isInComment_)
			{
				if (nextLine_.compare(i, 2, "*/") == 0)
				{
					isInComment_ = false;
					++i;
				}
				continue;
			}
			if (nextLine_[i] == '\\')
			{
				++i;
				continue;
			}
			if (isInQuote_)
			{
				if (nextLine_[i] == quoteChar_)
					isInQuote_ = false;
				continue;
			}

			if (nextLine_[i] == '"'
			        || (nextLine_[i] == '\'' && !isDigitSeparator(nextLine_, i)))
			{
				isInQuote_ = true;
				quoteChar_ = nextLine_[i];
				continue;
			}
			if (nextLine_.compare(i, 2, "//") == 0)
			{
				i = nextLine_.length();
				continue;
			}
			// handle preprocessor statement
			if (nextLine_[i] == '#')
			{
				std::string preproc = ASBeautifier::extractPreprocessorStatement(nextLine_);
				if (preproc.length() >= 2 && preproc.substr(0, 2) == "if") // #if, #ifdef, #ifndef
				{
					numBlockIndents += 1;
					isInIndentableBlock = true;
					// flag first preprocessor conditional for header include guard check
					if (!processedFirstConditional)
					{
						processedFirstConditional = true;
						isFirstPreprocConditional = true;
						if (isNDefPreprocStatement(nextLine_, preproc))
							isPotentialHeaderGuard = true;
					}
				}
				else if (preproc == "endif")
				{
					if (numBlockIndents > 0)
						numBlockIndents -= 1;
					// must exit BOTH loops
					if (numBlockIndents == 0)
						goto EndOfWhileLoop;
				}
				else if (preproc == "define")
				{
					if (nextLine_[nextLine_.length() - 1] == '\\')
						blockContainsDefineContinuation = true;
					// check for potential header include guards
					else if (isPotentialHeaderGuard && numBlockIndents == 1)
						isPotentialHeaderGuard2 = true;
				}
				i = nextLine_.length();
				continue;
			}
			// handle exceptions
			if (nextLine_[i] == '{' || nextLine_[i] == '}')
				blockContainsBraces = true;
			else if (nextLine_[i] == '(')
				++lineParenCount;
			else if (nextLine_[i] == ')')
				--lineParenCount;
			else if (nextLine_[i] == ':')
			{
				// check for '::'
				if (nextLine_.length() > i + 1 && nextLine_[i + 1] == ':')
					++i;
				else
					isInClassConstructor = true;
			}

			// bypass unnecessary parsing - must exit BOTH loops
			if (blockContainsBraces || isInClassConstructor || blockContainsDefineContinuation)
				goto EndOfWhileLoop;
		}	// end of for loop, end of line
		if (lineParenCount != 0)
			break;
	}	// end of while loop
EndOfWhileLoop:
	preprocBlockEnd = sourceIterator->tellg();
	if (preprocBlockEnd < 0)
		preprocBlockEnd = sourceIterator->getStreamLength();
	if (blockContainsBraces
	        || isInClassConstructor
	        || blockContainsDefineContinuation
	        || lineParenCount != 0
	        || numBlockIndents != 0)
		isInIndentableBlock = false;
	// find next executable instruction
	// this WILL RESET the get pointer
	std::string nextText = peekNextText("", false, stream);
	// bypass header include guards
	if (isFirstPreprocConditional)
	{
		isFirstPreprocConditional = false;
		if (nextText.empty() && isPotentialHeaderGuard2)
		{
			isInIndentableBlock = false;
			preprocBlockEnd = 0;
		}
	}
	// this allows preprocessor blocks within this block to be indented
	if (!isInIndentableBlock)
		preprocBlockEnd = 0;
	// peekReset() is done by previous peekNextText()
	return isInIndentableBlock;
}

bool ASFormatter::isNDefPreprocStatement(std::string_view nextLine_, std::string_view preproc) const
{
	if (preproc == "ifndef")
		return true;
	// check for '!defined'
	if (preproc == "if")
	{
		size_t i = nextLine_.find('!');
		if (i == std::string::npos)
			return false;
		i = nextLine_.find_first_not_of(" \t", ++i);
		if (i != std::string::npos && nextLine_.compare(i, 7, "defined") == 0)
			return true;
	}
	return false;
}

/**
 * Check to see if this is an EXEC SQL statement.
 *
 * @param line          a reference to the line to indent.
 * @param index         the current line index.
 * @return              true if the statement is EXEC SQL.
 */
bool ASFormatter::isExecSQL(std::string_view line, size_t index) const
{
	if (line[index] != 'e' && line[index] != 'E')	// quick check to reject most
		return false;
	std::string_view word;
	if (isCharPotentialHeader(line, index))
		word = getCurrentWord(line, index);
	for (char character : word)
		character = (char) toupper(character);
	if (word != "EXEC")
		return false;
	size_t index2 = index + word.length();
	index2 = line.find_first_not_of(" \t", index2);
	if (index2 == std::string::npos)
		return false;

	std::string_view word2;

	if (isCharPotentialHeader(line, index2))
		word2 = getCurrentWord(line, index2);
	for (char character : word2)
		character = (char) toupper(character);
	if (word2 != "SQL")
		return false;
	return true;
}

/**
 * The continuation lines must be adjusted so the leading spaces
 *     is equivalent to the text on the opening line.
 *
 * Updates currentLine and charNum.
 */
void ASFormatter::trimContinuationLine()
{
	size_t len = currentLine.length();
	size_t tabSize = getTabLength();
	charNum = 0;

	if (leadingSpaces > 0 && len > 0)
	{
		size_t i;
		size_t continuationIncrementIn = 0;
		for (i = 0; (i < len) && (i + continuationIncrementIn < leadingSpaces); i++)
		{
			if (!std::isblank(currentLine[i]))		// don't delete any text
			{
				if (i < continuationIncrementIn)
					leadingSpaces = i + tabIncrementIn;
				continuationIncrementIn = tabIncrementIn;
				break;
			}
			if (currentLine[i] == '\t')
				continuationIncrementIn += tabSize - 1 - ((continuationIncrementIn + i) % tabSize);
		}

		if ((int) continuationIncrementIn == tabIncrementIn)
			charNum = i;
		else
		{
			// build a new line with the equivalent leading chars
			std::string newLine;
			int leadingChars = 0;
			if ((int) leadingSpaces > tabIncrementIn)
				leadingChars = leadingSpaces - tabIncrementIn;
			newLine.append(leadingChars, ' ');
			newLine.append(currentLine, i, len - i);
			currentLine = newLine;
			charNum = leadingChars;
			if (currentLine.empty())
				currentLine = std::string(" ");        // a null is inserted if this is not done
		}
		if (i >= len)
			charNum = 0;
	}
}

/**
 * Determine if a header is a closing header
 *
 * @return      true if the header is a closing header.
 */
bool ASFormatter::isClosingHeader(const std::string* header) const
{
	return (header == &ASResource::AS_ELSE
	        || header == &ASResource::AS_CATCH
	        || header == &ASResource::AS_FINALLY);
}

/**
 * Determine if a * following a closing paren is immediately.
 * after a cast. If so it is a deference and not a multiply.
 * e.g. "(int*) *ptr" is a deference.
 */
bool ASFormatter::isImmediatelyPostCast() const
{
	assert(previousNonWSChar == ')' && currentChar == '*');
	// find preceding closing paren on currentLine or readyFormattedLine
	std::string line;		// currentLine or readyFormattedLine
	size_t paren = currentLine.rfind(')', charNum);
	if (paren != std::string::npos)
		line = currentLine;
	// if not on currentLine it must be on the previous line
	else
	{
		line = readyFormattedLine;
		paren = line.rfind(')');
		if (paren == std::string::npos)
			return false;
	}
	if (paren == 0)
		return false;

	// find character preceding the closing paren
	size_t lastChar = line.find_last_not_of(" \t", paren - 1);
	if (lastChar == std::string::npos)
		return false;
	// check for pointer cast
	if (line[lastChar] == '*')
		return true;
	return false;
}

/**
 * Determine if a < is a template definition or instantiation.
 * Sets the class variables isInTemplate and templateDepth.
 */
void ASFormatter::checkIfTemplateOpener()
{
	assert(!isInTemplate && currentChar == '<');

	// find first char after the '<' operators
	size_t firstChar = currentLine.find_first_not_of("< \t", charNum);
	if (firstChar == std::string::npos
	        || currentLine[firstChar] == '=')
	{
		// this is not a template -> leave...
		isInTemplate = false;
		return;
	}

	bool isFirstLine = true;
	int parenDepth_ = 0;
	int maxTemplateDepth = 0;
	templateDepth = 0;
	std::string nextLine_ = currentLine.substr(charNum);
	ASPeekStream stream(sourceIterator);

	// find the angle braces, bypassing all comments and quotes.
	bool isInComment_ = false;
	bool isInQuote_ = false;
	char quoteChar_ = ' ';
	while (stream.hasMoreLines() || isFirstLine)
	{
		if (isFirstLine)
			isFirstLine = false;
		else
			nextLine_ = stream.peekNextLine();
		// parse the line
		for (size_t i = 0; i < nextLine_.length(); i++)
		{
			char currentChar_ = nextLine_[i];
			if (std::isblank(currentChar_))
				continue;
			if (nextLine_.compare(i, 2, "/*") == 0)
				isInComment_ = true;
			if (isInComment_)
			{
				if (nextLine_.compare(i, 2, "*/") == 0)
				{
					isInComment_ = false;
					++i;
				}
				continue;
			}
			if (currentChar_ == '\\')
			{
				++i;
				continue;
			}

			if (isInQuote_)
			{
				if (currentChar_ == quoteChar_)
					isInQuote_ = false;
				continue;
			}

			if (currentChar_ == '"'
			        || (currentChar_ == '\'' && !isDigitSeparator(nextLine_, i)))
			{
				isInQuote_ = true;
				quoteChar_ = currentChar_;
				continue;
			}
			if (nextLine_.compare(i, 2, "//") == 0)
			{
				i = nextLine_.length();
				continue;
			}

			// not in a comment or quote
			if (currentChar_ == '<')
			{
				++templateDepth;
				++maxTemplateDepth;
				continue;
			}
			if (currentChar_ == '>')
			{
				--templateDepth;
				if (templateDepth == 0)
				{
					if (parenDepth_ == 0)
					{
						// this is a template!
						// gl85 + sf585
						isInTemplate = !isInStruct;
						templateDepth = maxTemplateDepth;
					}
					return;
				}
				continue;
			}
			if (currentChar_ == '(' || currentChar_ == ')')
			{
				if (currentChar_ == '(')
					++parenDepth_;
				else
					--parenDepth_;
				if (parenDepth_ >= 0)
					continue;
				// this is not a template -> leave...
				isInTemplate = false;
				templateDepth = 0;
				return;
			}
			if (nextLine_.compare(i, 2, ASResource::AS_AND) == 0
			        || nextLine_.compare(i, 2, ASResource::AS_OR) == 0)
			{
				// this is not a template -> leave...
				isInTemplate = false;
				templateDepth = 0;
				return;
			}

			if (currentChar_ == ','  // comma,     e.g. A<int, char>
			        || currentChar_ == '&'    // reference, e.g. A<int&>
			        || currentChar_ == '*'    // pointer,   e.g. A<int*>
			        || currentChar_ == '^'    // C++/CLI managed pointer, e.g. A<int^>
			        || currentChar_ == ':'    // ::,        e.g. std::string
			        || currentChar_ == '='    // assign     e.g. default parameter
			        || currentChar_ == '['    // []         e.g. std::string[]
			        || currentChar_ == ']'    // []         e.g. std::string[]
			        || currentChar_ == '('    // (...)      e.g. function definition
			        || currentChar_ == ')'    // (...)      e.g. function definition
			        || (isJavaStyle() && currentChar_ == '?')   // Java wildcard
			   )
			{
				continue;
			}
			if (!isLegalNameChar(currentChar_))
			{
				// this is not a template -> leave...
				isInTemplate = false;
				templateDepth = 0;
				return;
			}
			std::string_view name = getCurrentWord(nextLine_, i);
			i += name.length() - 1;
		}	// end for loop
	}	// end while loop
}

void ASFormatter::updateFormattedLineSplitPoints(char appendedChar)
{
	assert(maxCodeLength != std::string::npos);
	assert(!formattedLine.empty());

	if (!isOkToSplitFormattedLine())
		return;

	char nextChar = peekNextChar();

	// don't split before an end of line comment
	if (nextChar == '/')
		return;

	// don't split before or after a brace
	if (appendedChar == '{' || appendedChar == '}'
	        || previousNonWSChar == '{' || previousNonWSChar == '}'
	        || nextChar == '{' || nextChar == '}'
	        || currentChar == '{' || currentChar == '}')	// currentChar tests for an appended brace
		return;

	// don't split before or after a block paren
	if (appendedChar == '[' || appendedChar == ']'
	        || previousNonWSChar == '['
	        || nextChar == '[' || nextChar == ']')
		return;

	if (std::isblank(appendedChar))
	{
		if (nextChar != ')'						// space before a closing paren
		        && nextChar != '('				// space before an opening paren
		        && nextChar != '/'				// space before a comment
		        && nextChar != ':'				// space before a colon
		        && currentChar != ')'			// appended space before and after a closing paren
		        && currentChar != '('			// appended space before and after a opening paren
		        && previousNonWSChar != '('		// decided at the '('
		        // don't break before a pointer or reference aligned to type
		        && !(nextChar == '*'
		             && !isCharPotentialOperator(previousNonWSChar)
		             && pointerAlignment == PTR_ALIGN_TYPE)
		        && !(nextChar == '&'
		             && !isCharPotentialOperator(previousNonWSChar)
		             && (referenceAlignment == REF_ALIGN_TYPE
		                 || (referenceAlignment == REF_SAME_AS_PTR && pointerAlignment == PTR_ALIGN_TYPE)))
		   )
		{
			if (formattedLine.length() - 1 <= maxCodeLength)
				maxWhiteSpace = formattedLine.length() - 1;
			else
				maxWhiteSpacePending = formattedLine.length() - 1;
		}
	}
	// unpadded closing parens may split after the paren (counts as whitespace)
	else if (appendedChar == ')')
	{
		if (nextChar != ')'
		        && nextChar != ' '
		        && nextChar != ';'
		        && nextChar != ','
		        && nextChar != '.'
		        && !(nextChar == '-' && pointerSymbolFollows()))	// check for ->
		{
			if (formattedLine.length() <= maxCodeLength)
				maxWhiteSpace = formattedLine.length();
			else
				maxWhiteSpacePending = formattedLine.length();
		}
	}
	// unpadded commas may split after the comma
	else if (appendedChar == ',')
	{
		if (formattedLine.length() <= maxCodeLength)
			maxComma = formattedLine.length();
		else
			maxCommaPending = formattedLine.length();
	}
	else if (appendedChar == '(')
	{
		if (nextChar != ')' && nextChar != '(' && nextChar != '"' && nextChar != '\'')
		{
			// if follows an operator break before
			size_t parenNum;
			if (previousNonWSChar != ' ' && isCharPotentialOperator(previousNonWSChar))
				parenNum = formattedLine.length() - 1;
			else
				parenNum = formattedLine.length();
			if (formattedLine.length() <= maxCodeLength)
				maxParen = parenNum;
			else
				maxParenPending = parenNum;
		}
	}
	else if (appendedChar == ';')
	{
		if (nextChar != ' '  && nextChar != '}' && nextChar != '/')	// check for following comment
		{
			if (formattedLine.length() <= maxCodeLength)
				maxSemi = formattedLine.length();
			else
				maxSemiPending = formattedLine.length();
		}
	}
}

void ASFormatter::updateFormattedLineSplitPointsOperator(std::string_view sequence)
{
	assert(maxCodeLength != std::string::npos);
	assert(!formattedLine.empty());

	if (!isOkToSplitFormattedLine())
		return;

	char nextChar = peekNextChar();

	// don't split before an end of line comment
	if (nextChar == '/')
		return;

	// check for logical conditional
	if (sequence == "||" || sequence == "&&" || sequence == "or" || sequence == "and")
	{
		if (shouldBreakLineAfterLogical)
		{
			if (formattedLine.length() <= maxCodeLength)
				maxAndOr = formattedLine.length();
			else
				maxAndOrPending = formattedLine.length();
		}
		else
		{
			// adjust for leading space in the sequence
			size_t sequenceLength = sequence.length();
			if (formattedLine.length() > sequenceLength
			        && std::isblank(formattedLine[formattedLine.length() - sequenceLength - 1]))
				sequenceLength++;
			if (formattedLine.length() - sequenceLength <= maxCodeLength)
				maxAndOr = formattedLine.length() - sequenceLength;
			else
				maxAndOrPending = formattedLine.length() - sequenceLength;
		}
	}
	// comparison operators will split after the operator (counts as whitespace)
	else if (sequence == "==" || sequence == "!=" || sequence == ">=" || sequence == "<=")
	{
		if (formattedLine.length() <= maxCodeLength)
			maxWhiteSpace = formattedLine.length();
		else
			maxWhiteSpacePending = formattedLine.length();
	}
	// unpadded operators that will split BEFORE the operator (counts as whitespace)
	else if (sequence == "+" || sequence == "-" || sequence == "?")
	{
		if (charNum > 0
		        && !(sequence == "+" && isInExponent())
		        && !(sequence == "-"  && isInExponent())
		        && (isLegalNameChar(currentLine[charNum - 1])
		            || currentLine[charNum - 1] == ')'
		            || currentLine[charNum - 1] == ']'
		            || currentLine[charNum - 1] == '\"'))
		{
			if (formattedLine.length() - 1 <= maxCodeLength)
				maxWhiteSpace = formattedLine.length() - 1;
			else
				maxWhiteSpacePending = formattedLine.length() - 1;
		}
	}
	// unpadded operators that will USUALLY split AFTER the operator (counts as whitespace)
	else if (sequence == "=" || sequence == ":")
	{
		// split BEFORE if the line is too long
		// do NOT use <= here, must allow for a brace attached to an array
		size_t splitPoint = 0;
		if (formattedLine.length() < maxCodeLength)
			splitPoint = formattedLine.length();
		else
			splitPoint = formattedLine.length() - 1;
		// padded or unpadded arrays
		if (previousNonWSChar == ']')
		{
			if (formattedLine.length() - 1 <= maxCodeLength)
				maxWhiteSpace = splitPoint;
			else
				maxWhiteSpacePending = splitPoint;
		}
		else if (charNum > 0
		         && (isLegalNameChar(currentLine[charNum - 1])
		             || currentLine[charNum - 1] == ')'
		             || currentLine[charNum - 1] == ']'))
		{
			if (formattedLine.length() <= maxCodeLength)
				maxWhiteSpace = splitPoint;
			else
				maxWhiteSpacePending = splitPoint;
		}
	}
}

/**
 * Update the split point when a pointer or reference is formatted.
 * The argument is the maximum index of the last whitespace character.
 */
void ASFormatter::updateFormattedLineSplitPointsPointerOrReference(size_t index)
{
	assert(maxCodeLength != std::string::npos);
	assert(!formattedLine.empty());
	assert(index < formattedLine.length());

	if (!isOkToSplitFormattedLine())
		return;

	if (index < maxWhiteSpace)		// just in case
		return;

	if (index <= maxCodeLength)
		maxWhiteSpace = index;
	else
		maxWhiteSpacePending = index;
}

bool ASFormatter::isOkToSplitFormattedLine()
{
	assert(maxCodeLength != std::string::npos);
	// Is it OK to split the line?
	if (shouldKeepLineUnbroken
	        || isInLineComment
	        || isInComment
	        || isInQuote
	        || isInCase
	        || isInPreprocessor
	        || isInExecSQL
	        || isInAsm || isInAsmOneLine || isInAsmBlock
	        || isInTemplate)
		return false;

	if (!isOkToBreakBlock(braceTypeStack->back()) && currentChar != '{')
	{
		shouldKeepLineUnbroken = true;
		clearFormattedLineSplitPoints();
		return false;
	}
	if (isBraceType(braceTypeStack->back(), ARRAY_TYPE))
	{
		shouldKeepLineUnbroken = true;
		if (!isBraceType(braceTypeStack->back(), ARRAY_NIS_TYPE))
			clearFormattedLineSplitPoints();
		return false;
	}
	return true;
}

/* This is called if the option maxCodeLength is set.
 */
void ASFormatter::testForTimeToSplitFormattedLine()
{
	//	DO NOT ASSERT maxCodeLength HERE
	// should the line be split
	if (formattedLine.length() > maxCodeLength && !isLineReady)
	{
		size_t splitPoint = findFormattedLineSplitPoint();
		if (splitPoint > 0 && splitPoint < formattedLine.length())
		{
			std::string splitLine = formattedLine.substr(splitPoint);
			formattedLine = formattedLine.substr(0, splitPoint);
			breakLine(true);
			formattedLine = splitLine;
			// if break-blocks is requested and this is a one-line statement
			std::string nextWord = ASBeautifier::getNextWord(currentLine, charNum - 1);
			if (isAppendPostBlockEmptyLineRequested
			        && (nextWord == "break" || nextWord == "continue"))
			{
				isAppendPostBlockEmptyLineRequested = false;
				isPrependPostBlockEmptyLineRequested = true;
			}
			else
				isPrependPostBlockEmptyLineRequested = false;
			// adjust max split points
			maxAndOr = (maxAndOr > splitPoint) ? (maxAndOr - splitPoint) : 0;
			maxSemi = (maxSemi > splitPoint) ? (maxSemi - splitPoint) : 0;
			maxComma = (maxComma > splitPoint) ? (maxComma - splitPoint) : 0;
			maxParen = (maxParen > splitPoint) ? (maxParen - splitPoint) : 0;
			maxWhiteSpace = (maxWhiteSpace > splitPoint) ? (maxWhiteSpace - splitPoint) : 0;
			if (maxSemiPending > 0)
			{
				maxSemi = (maxSemiPending > splitPoint) ? (maxSemiPending - splitPoint) : 0;
				maxSemiPending = 0;
			}
			if (maxAndOrPending > 0)
			{
				maxAndOr = (maxAndOrPending > splitPoint) ? (maxAndOrPending - splitPoint) : 0;
				maxAndOrPending = 0;
			}
			if (maxCommaPending > 0)
			{
				maxComma = (maxCommaPending > splitPoint) ? (maxCommaPending - splitPoint) : 0;
				maxCommaPending = 0;
			}
			if (maxParenPending > 0)
			{
				maxParen = (maxParenPending > splitPoint) ? (maxParenPending - splitPoint) : 0;
				maxParenPending = 0;
			}
			if (maxWhiteSpacePending > 0)
			{
				maxWhiteSpace = (maxWhiteSpacePending > splitPoint) ? (maxWhiteSpacePending - splitPoint) : 0;
				maxWhiteSpacePending = 0;
			}
			// don't allow an empty formatted line
			size_t firstText = formattedLine.find_first_not_of(" \t");
			if (firstText == std::string::npos && !formattedLine.empty())
			{
				formattedLine.erase();
				clearFormattedLineSplitPoints();
				if (std::isblank(currentChar))
					for (size_t i = charNum + 1; i < currentLine.length() && std::isblank(currentLine[i]); i++)
						goForward(1);
			}
			else if (firstText > 0)
			{
				formattedLine.erase(0, firstText);
				maxSemi = (maxSemi > firstText) ? (maxSemi - firstText) : 0;
				maxAndOr = (maxAndOr > firstText) ? (maxAndOr - firstText) : 0;
				maxComma = (maxComma > firstText) ? (maxComma - firstText) : 0;
				maxParen = (maxParen > firstText) ? (maxParen - firstText) : 0;
				maxWhiteSpace = (maxWhiteSpace > firstText) ? (maxWhiteSpace - firstText) : 0;
			}
			// reset formattedLineCommentNum
			if (formattedLineCommentNum != std::string::npos)
			{
				formattedLineCommentNum = formattedLine.find("//");
				if (formattedLineCommentNum == std::string::npos)
					formattedLineCommentNum = formattedLine.find("/*");
			}
		}
	}
}

size_t ASFormatter::findFormattedLineSplitPoint() const
{
	assert(maxCodeLength != std::string::npos);
	// determine where to split
	size_t minCodeLength = 10;
	size_t splitPoint = 0;
	splitPoint = maxSemi;
	if (maxAndOr >= minCodeLength)
		splitPoint = maxAndOr;
	if (splitPoint < minCodeLength)
	{
		splitPoint = maxWhiteSpace;
		// use maxParen instead if it is long enough
		if (maxParen > splitPoint
		        || maxParen >= maxCodeLength * .7)
			splitPoint = maxParen;
		// use maxComma instead if it is long enough
		// increasing the multiplier causes more splits at whitespace
		if (maxComma > splitPoint
		        || maxComma >= maxCodeLength * .3)
			splitPoint = maxComma;
	}
	// replace split point with first available break point
	if (splitPoint < minCodeLength)
	{
		splitPoint = std::string::npos;
		if (maxSemiPending > 0 && maxSemiPending < splitPoint)
			splitPoint = maxSemiPending;
		if (maxAndOrPending > 0 && maxAndOrPending < splitPoint)
			splitPoint = maxAndOrPending;
		if (maxCommaPending > 0 && maxCommaPending < splitPoint)
			splitPoint = maxCommaPending;
		if (maxParenPending > 0 && maxParenPending < splitPoint)
			splitPoint = maxParenPending;
		if (maxWhiteSpacePending > 0 && maxWhiteSpacePending < splitPoint)
			splitPoint = maxWhiteSpacePending;
		if (splitPoint == std::string::npos)
			splitPoint = 0;
	}
	// if remaining line after split is too long
	else if (formattedLine.length() - splitPoint > maxCodeLength)
	{
		// if end of the currentLine, find a new split point
		size_t newCharNum;
		if (!std::isblank(currentChar) && isCharPotentialHeader(currentLine, charNum))
			newCharNum = getCurrentWord(currentLine, charNum).length() + charNum;
		else
			newCharNum = charNum + 2;
		if (newCharNum + 1 > currentLine.length())
		{
			// don't move splitPoint from before a conditional to after
			if (maxWhiteSpace > splitPoint + 3)
				splitPoint = maxWhiteSpace;
			if (maxParen > splitPoint)
				splitPoint = maxParen;
		}
	}

	return splitPoint;
}

void ASFormatter::clearFormattedLineSplitPoints()
{
	maxSemi = 0;
	maxAndOr = 0;
	maxComma = 0;
	maxParen = 0;
	maxWhiteSpace = 0;
	maxSemiPending = 0;
	maxAndOrPending = 0;
	maxCommaPending = 0;
	maxParenPending = 0;
	maxWhiteSpacePending = 0;
}

/**
 * Check if a pointer symbol (->) follows on the currentLine.
 */
bool ASFormatter::pointerSymbolFollows() const
{
	size_t peekNum = currentLine.find_first_not_of(" \t", charNum + 1);
	if (peekNum == std::string::npos || currentLine.compare(peekNum, 2, "->") != 0)
		return false;
	return true;
}

/**
 * Compute the input checksum.
 * This is called as an assert so it for is debug config only
 */
bool ASFormatter::computeChecksumIn(std::string_view currentLine_)
{
	for (const char character : currentLine_)
		if (!std::isblank(character))
			checksumIn += character;
	return true;
}

/**
 * Adjust the input checksum for deleted chars.
 * This is called as an assert so it for is debug config only
 */
bool ASFormatter::adjustChecksumIn(int adjustment)
{
	checksumIn += adjustment;
	return true;
}

/**
 * get the value of checksumIn for unit testing
 *
 * @return   checksumIn.
 */
size_t ASFormatter::getChecksumIn() const
{
	return checksumIn;
}

/**
 * Compute the output checksum.
 * This is called as an assert so it is for debug config only
 */
bool ASFormatter::computeChecksumOut(std::string_view beautifiedLine)
{
	for (const char character : beautifiedLine)
		if (!std::isblank(character))
			checksumOut += character;
	return true;
}

/**
 * Return isLineReady for the final check at end of file.
 */
bool ASFormatter::getIsLineReady() const
{
	return isLineReady;
}

/**
 * get the value of checksumOut for unit testing
 *
 * @return   checksumOut.
 */
size_t ASFormatter::getChecksumOut() const
{
	return checksumOut;
}

/**
 * Return the difference in checksums.
 * If zero all is okay.
 */
int ASFormatter::getChecksumDiff() const
{
	return checksumOut - checksumIn;
}

// for unit testing
int ASFormatter::getFormatterFileType() const
{
	return formatterFileType;
}

// Check if an operator follows the next word.
// The next word must be a legal name.
const std::string* ASFormatter::getFollowingOperator() const
{
	// find next word
	size_t nextNum = currentLine.find_first_not_of(" \t", charNum + 1);
	if (nextNum == std::string::npos)
		return nullptr;

	if (!isLegalNameChar(currentLine[nextNum]))
		return nullptr;

	// bypass next word and following spaces
	while (nextNum < currentLine.length())
	{
		if (!isLegalNameChar(currentLine[nextNum])
		        && !std::isblank(currentLine[nextNum]))
			break;
		nextNum++;
	}

	if (nextNum >= currentLine.length()
	        || !isCharPotentialOperator(currentLine[nextNum])
	        || currentLine[nextNum] == '/')		// comment
		return nullptr;

	const std::string* newOperator = ASBase::findOperator(currentLine, nextNum, operators);
	return newOperator;
}

// Check following data to determine if the current character is an array operator.
bool ASFormatter::isArrayOperator() const
{
	assert(currentChar == '*' || currentChar == '&' || currentChar == '^');
	assert(isBraceType(braceTypeStack->back(), ARRAY_TYPE));

	// find next word
	size_t nextNum = currentLine.find_first_not_of(" \t", charNum + 1);
	if (nextNum == std::string::npos)
		return false;

	if (!isLegalNameChar(currentLine[nextNum]))
		return false;

	// bypass next word and following spaces
	while (nextNum < currentLine.length())
	{
		if (!isLegalNameChar(currentLine[nextNum])
		        && !std::isblank(currentLine[nextNum]))
			break;
		nextNum++;
	}

	// check for characters that indicate an operator
	if (currentLine[nextNum] == ','
	        || currentLine[nextNum] == '}'
	        || currentLine[nextNum] == ')'
	        || currentLine[nextNum] == '(')
		return true;
	return false;
}

// Reset the flags that indicate various statement information.
void ASFormatter::resetEndOfStatement()
{
	foundQuestionMark = false;
	foundNamespaceHeader = false;
	foundClassHeader = false;
	foundStructHeader = false;
	foundInterfaceHeader = false;
	foundPreDefinitionHeader = false;
	foundPreCommandHeader = false;
	foundPreCommandMacro = false;
	foundTrailingReturnType = false;
	foundCastOperator = false;
	isInPotentialCalculation = false;
	isSharpAccessor = false;
	isSharpDelegate = false;
	isInObjCMethodDefinition = false;
	isImmediatelyPostObjCMethodPrefix = false;
	isInObjCReturnType = false;
	isInObjCParam = false;
	isInObjCInterface = false;
	isInObjCSelector = false;
	isInEnum = false;
	isInExternC = false;
	elseHeaderFollowsComments = false;
	returnTypeChecked = false;
	nonInStatementBrace = 0;
	while (!questionMarkStack->empty())
		questionMarkStack->pop_back();
}

// Find the colon alignment for Objective-C method definitions and method calls.
int ASFormatter::findObjCColonAlignment() const
{
	assert(currentChar == '+' || currentChar == '-' || currentChar == '[');
	assert(getAlignMethodColon());

	bool isFirstLine = true;
	bool haveFirstColon = false;
	bool foundMethodColon = false;
	bool isInComment_ = false;
	bool isInQuote_ = false;
	bool haveTernary = false;
	char quoteChar_ = ' ';
	int  sqBracketCount = 0;
	int  colonAdjust = 0;
	int  colonAlign = 0;
	std::string nextLine_ = currentLine;
	ASPeekStream stream(sourceIterator);

	// peek next line
	while (sourceIterator->hasMoreLines() || isFirstLine)
	{
		if (!isFirstLine)
			nextLine_ = stream.peekNextLine();
		// parse the line
		haveFirstColon = false;
		nextLine_ = ASBeautifier::trim(nextLine_);
		for (size_t i = 0; i < nextLine_.length(); i++)
		{
			if (std::isblank(nextLine_[i]))
				continue;
			if (nextLine_.compare(i, 2, "/*") == 0)
				isInComment_ = true;
			if (isInComment_)
			{
				if (nextLine_.compare(i, 2, "*/") == 0)
				{
					isInComment_ = false;
					++i;
				}
				continue;
			}
			if (nextLine_[i] == '\\')
			{
				++i;
				continue;
			}
			if (isInQuote_)
			{
				if (nextLine_[i] == quoteChar_)
					isInQuote_ = false;
				continue;
			}

			if (nextLine_[i] == '"'
			        || (nextLine_[i] == '\'' && !isDigitSeparator(nextLine_, i)))
			{
				isInQuote_ = true;
				quoteChar_ = nextLine_[i];
				continue;
			}
			if (nextLine_.compare(i, 2, "//") == 0)
			{
				i = nextLine_.length();
				continue;
			}
			// process the current char
			if ((nextLine_[i] == '{' && (currentChar == '-' || currentChar == '+'))
			        || nextLine_[i] == ';')
				goto EndOfWhileLoop;       // end of method definition
			if (nextLine_[i] == ']')
			{
				--sqBracketCount;
				if (sqBracketCount == 0)
					goto EndOfWhileLoop;   // end of method call
			}
			if (nextLine_[i] == '[')
				++sqBracketCount;
			if (isFirstLine)	 // colon align does not include the first line
				continue;
			if (sqBracketCount > 1)
				continue;
			if (haveFirstColon)  // multiple colons per line
				continue;
			if (nextLine_[i] == '?')
			{
				haveTernary = true;
				continue;
			}
			// compute colon adjustment
			if (nextLine_[i] == ':')
			{
				if (haveTernary)
				{
					haveTernary = false;
					continue;
				}
				haveFirstColon = true;
				foundMethodColon = true;
				if (isObjCStyle() && shouldPadMethodColon)
				{
					int spacesStart;
					for (spacesStart = i; spacesStart > 0; spacesStart--)
						if (!std::isblank(nextLine_[spacesStart - 1]))
							break;
					int spaces = i - spacesStart;
					if (objCColonPadMode == COLON_PAD_ALL || objCColonPadMode == COLON_PAD_BEFORE)
						colonAdjust = 1 - spaces;
					else if (objCColonPadMode == COLON_PAD_NONE || objCColonPadMode == COLON_PAD_AFTER)
						colonAdjust = 0 - spaces;
				}
				// compute alignment
				int colonPosition = i + colonAdjust;
				if (colonPosition > colonAlign)
					colonAlign = colonPosition;
			}
		}	// end of for loop
		isFirstLine = false;
	}	// end of while loop
EndOfWhileLoop:
	if (!foundMethodColon)
		colonAlign = -1;
	return colonAlign;
}

// pad an Objective-C method colon
void ASFormatter::padObjCMethodColon()
{
	assert(currentChar == ':');
	int commentAdjust = 0;
	char nextChar = peekNextChar();
	if (objCColonPadMode == COLON_PAD_NONE
	        || objCColonPadMode == COLON_PAD_AFTER
	        || nextChar == ')')
	{
		// remove spaces before
		for (int i = formattedLine.length() - 1; (i > -1) && std::isblank(formattedLine[i]); i--)
		{
			formattedLine.erase(i);
			--commentAdjust;
		}
	}
	else
	{
		// pad space before
		for (int i = formattedLine.length() - 1; (i > 0) && std::isblank(formattedLine[i]); i--)
			if (std::isblank(formattedLine[i - 1]))
			{
				formattedLine.erase(i);
				--commentAdjust;
			}
		if (!formattedLine.empty())
		{
			appendSpacePad();
			formattedLine.back() = ' ';  // convert any tab to space
		}
	}
	if (objCColonPadMode == COLON_PAD_NONE
	        || objCColonPadMode == COLON_PAD_BEFORE
	        || nextChar == ')')
	{
		// remove spaces after
		size_t nextText = currentLine.find_first_not_of(" \t", charNum + 1);
		if (nextText == std::string::npos)
			nextText = currentLine.length();
		int spaces = nextText - charNum - 1;
		if (spaces > 0)
		{
			// do not use goForward here
			currentLine.erase(charNum + 1, spaces);
			spacePadNum -= spaces;
		}
	}
	else
	{
		// pad space after
		size_t nextText = currentLine.find_first_not_of(" \t", charNum + 1);
		if (nextText == std::string::npos)
			nextText = currentLine.length();
		int spaces = nextText - charNum - 1;
		if (spaces == 0)
		{
			currentLine.insert(charNum + 1, 1, ' ');
			spacePadNum += 1;
		}
		else if (spaces > 1)
		{
			// do not use goForward here
			currentLine.erase(charNum + 1, spaces - 1);
			currentLine[charNum + 1] = ' ';  // convert any tab to space
			spacePadNum -= spaces - 1;
		}
	}
	spacePadNum += commentAdjust;
}

// Remove the leading '*' from a comment line and indent to the next tab.
void ASFormatter::stripCommentPrefix()
{
	int firstChar = formattedLine.find_first_not_of(" \t");
	if (firstChar < 0)
		return;

	if (isInCommentStartLine)
	{
		// comment opener must begin the line
		if (formattedLine.compare(firstChar, 2, "/*") != 0)
			return;
		int commentOpener = firstChar;
		// ignore single line comments
		int commentEnd = formattedLine.find("*/", firstChar + 2);
		if (commentEnd != -1)
			return;
		// first char after the comment opener must be at least one indent
		int followingText = formattedLine.find_first_not_of(" \t", commentOpener + 2);
		if (followingText < 0)
			return;
		if (formattedLine[followingText] == '*' || formattedLine[followingText] == '!')
			followingText = formattedLine.find_first_not_of(" \t", followingText + 1);
		if (followingText < 0)
			return;
		if (formattedLine[followingText] == '*')
			return;
		int indentLen = getIndentLength();
		int followingTextIndent = followingText - commentOpener;
		if (followingTextIndent < indentLen)
		{
			std::string stringToInsert(indentLen - followingTextIndent, ' ');
			formattedLine.insert(followingText, stringToInsert);
		}
		return;
	}
	// comment body including the closer
	if (formattedLine[firstChar] == '*')
	{
		if (formattedLine.compare(firstChar, 2, "*/") == 0)
		{
			// line starts with an end comment
			formattedLine = "*/";
		}
		else
		{
			// build a new line with one indent
			int secondChar = formattedLine.find_first_not_of(" \t", firstChar + 1);
			if (secondChar < 0)
			{
				adjustChecksumIn(-'*');
				formattedLine.erase();
				return;
			}
			if (formattedLine[secondChar] == '*')
				return;
			// replace the leading '*'
			int indentLen = getIndentLength();
			adjustChecksumIn(-'*');
			// second char must be at least one indent
			if (formattedLine.substr(0, secondChar).find('\t') != std::string::npos)
			{
				formattedLine.erase(firstChar, 1);
			}
			else
			{
				int spacesToInsert = 0;
				if (secondChar >= indentLen)
					spacesToInsert = secondChar;
				else
					spacesToInsert = indentLen;
				formattedLine = std::string(spacesToInsert, ' ') + formattedLine.substr(secondChar);
			}
			// remove a trailing '*'
			int lastChar = formattedLine.find_last_not_of(" \t");
			if (lastChar > -1 && formattedLine[lastChar] == '*')
			{
				adjustChecksumIn(-'*');
				formattedLine[lastChar] = ' ';
			}
		}
	}
	else
	{
		// first char not a '*'
		// first char must be at least one indent
		if (formattedLine.substr(0, firstChar).find('\t') == std::string::npos)
		{
			int indentLen = getIndentLength();
			if (firstChar < indentLen)
			{
				std::string stringToInsert(indentLen, ' ');
				formattedLine = stringToInsert + formattedLine.substr(firstChar);
			}
		}
	}
}

}   // end namespace astyle
