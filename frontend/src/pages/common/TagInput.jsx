import { useState } from 'react';
import { FiX } from 'react-icons/fi';

function normalizeTag(raw) {
  return raw.trim().toLowerCase().replace(/[^a-z0-9-]/g, '').slice(0, 30);
}

export function TagInput({ tags = [], onChange }) {
  const [input, setInput] = useState('');

  const addTag = (raw) => {
    const tag = normalizeTag(raw);
    if (tag && !tags.includes(tag)) {
      onChange([...tags, tag]);
    }
    setInput('');
  };

  const removeTag = (tag) => onChange(tags.filter(t => t !== tag));

  const handleKeyDown = (e) => {
    if (e.key === 'Enter' || e.key === ',') {
      e.preventDefault();
      addTag(input);
    } else if (e.key === 'Backspace' && !input && tags.length > 0) {
      removeTag(tags[tags.length - 1]);
    }
  };

  return (
    <div className="flex flex-wrap gap-1.5 p-2 border border-gray-300 rounded-lg focus-within:ring-2 focus-within:ring-indigo-500 focus-within:border-transparent min-h-[42px] cursor-text">
      {tags.map(tag => (
        <span key={tag} className="inline-flex items-center gap-1 px-2 py-0.5 bg-indigo-100 text-indigo-800 rounded-full text-sm font-medium">
          {tag}
          <button type="button" onClick={() => removeTag(tag)} className="hover:text-indigo-600 leading-none">
            <FiX size={12} />
          </button>
        </span>
      ))}
      <input
        type="text"
        value={input}
        onChange={e => setInput(e.target.value)}
        onKeyDown={handleKeyDown}
        onBlur={() => { if (input.trim()) addTag(input); }}
        placeholder={tags.length === 0 ? 'Add tags (Enter or comma to add)...' : ''}
        className="flex-1 min-w-[160px] outline-none text-sm bg-transparent"
      />
    </div>
  );
}
