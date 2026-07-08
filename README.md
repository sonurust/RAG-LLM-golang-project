# RAG-LLM

A small Go project implementing a REPL for a Retrieval-Augmented Generation (RAG) chat interface.

This repository is open source and available under the MIT License.

Author: Sonu Kumar
Contact:
- Email: [sonurust@gmail.com](mailto:sonurust@gmail.com)
- WhatsApp / Call / SMS: [+91 98106 59036](https://wa.me/919810659036)
- Instagram: [skbhati1992](https://www.instagram.com/skbhati1992)
- Telegram: [skbhati199](https://t.me/skbhati199)
- Facebook: [skbhati199](https://www.facebook.com/skbhati199)

## Getting Started

1. Copy or update the `.env` file with your OpenAI-compatible settings.
2. Run the REPL:

```bash
go run ./cmd/rag/
```

## Environment Variables

The project loads configuration from `.env` using `godotenv`.

```env
OPENAI_BASE_URL=https://api.openai.com/v1
OPENAI_API_KEY=your-openai-api-key
OPENAI_MODEL=gpt-4-mini
# SYSTEM_PROMPT_FILE=./system_prompt.txt
```

