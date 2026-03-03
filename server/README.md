# Elysium Registry Server

## Setup

1. Install dependencies:
```bash
pip install -r requirements.txt
```

2. Create `.env` file:
```bash
SUPABASE_URL=https://your-project.supabase.co
SUPABASE_ANON_KEY=your-anon-key
SUPABASE_SERVICE_KEY=your-service-key
```

3. Run the server:
```bash
uvicorn app.main:app --reload
```

## API Documentation

Access at `/docs` when running.