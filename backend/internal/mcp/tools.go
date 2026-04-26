package mcp

import "google.golang.org/genai"

// SearchWebTool — set web search tool for Gemini
var SearchWebTool = &genai.Tool{
	FunctionDeclarations: []*genai.FunctionDeclaration{
		{
			Name:        "search_web",
			Description: "Search the web for current information, news, or any topic not covered by the document context.",
			Parameters: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"query": {
						Type:        genai.TypeString,
						Description: "The search query to look up on the web",
					},
				},
				Required: []string{"query"},
			},
		},
	},
}

// SearchDocumentsTool — set document search tool for Gemini
var SearchDocumentsTool = &genai.Tool{
	FunctionDeclarations: []*genai.FunctionDeclaration{
		{
			Name:        "search_documents",
			Description: "Search the uploaded documents for this room to find relevant information from PDFs.",
			Parameters: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"query": {
						Type:        genai.TypeString,
						Description: "The search query to look up in the room documents",
					},
				},
				Required: []string{"query"},
			},
		},
	},
}
