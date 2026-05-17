export async function loginAPI(email, password) {
    const res = await fetch('/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, password }),
    });
    const text = await res.text();
    if (!res.ok) throw new Error(text.trim() || `HTTP ${res.status}`);
    return JSON.parse(text);
}

export async function getUserInfo(token) {
    const res = await fetch('/auth/me', {
        headers: { 'Authorization': `Bearer ${token}` },
    });
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    return res.json();
}

export async function registerAPI(name, email, password) {
    const res = await fetch('/auth/register', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name, email, password }),
    });
    if (!res.ok) {
        const text = await res.text();
        throw new Error(text.trim() || `HTTP ${res.status}`);
    }
}
