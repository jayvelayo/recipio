import requests
import time

prompt = """
Extract the recipe into JSON with this schema:
{
  "title": string,
  "ingredients": [{"name": string, "quantity": string}],
  "steps": [string]
}

Recipe:
French toast
2 eggs
1 cup milk
3 tsp parsley, minced

-. Crack open the egg
-. Add milk and egg into a bowl
- Add parsley if desired
-. That's it.

Return ONLY valid JSON.
"""

LLM_API_URL = 'http://localhost:11434/api/generate'

start = time.perf_counter()
r = requests.post(LLM_API_URL, json={
    'model': 'qwen3:4b',
    'prompt': prompt,
    'format': 'json',
    'stream': False,
    'think': False,
    'options': {'temperature': 0},
})
end = time.perf_counter()

print(f"Elapsed time: {end - start:.6f} seconds")
print(r.json()['response'])
