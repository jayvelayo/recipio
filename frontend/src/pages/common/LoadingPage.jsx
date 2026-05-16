
import { FiLoader } from 'react-icons/fi';
import { useState, useEffect } from 'react';

function LoadingPage() {
    return (
        <div className="flex items-center justify-center py-12">
            <div className="text-center">
                <FiLoader className="animate-spin w-10 h-10 text-indigo-600 mx-auto mb-4" />
                <p className="text-gray-600 text-lg font-medium">Loading...</p>
            </div>
        </div>
    )
}

export function SkeletonList({ rows = 5, delay = 150 }) {
    const [visible, setVisible] = useState(false);

    useEffect(() => {
        const t = setTimeout(() => setVisible(true), delay);
        return () => clearTimeout(t);
    }, [delay]);

    if (!visible) return <span className="sr-only">Loading...</span>;

    return (
        <div className="space-y-6" aria-busy="true">
            <span className="sr-only">Loading...</span>
            <div className="flex items-center justify-between">
                <div className="h-8 w-36 bg-gray-200 rounded-lg animate-pulse" />
                <div className="h-10 w-32 bg-gray-200 rounded-lg animate-pulse" />
            </div>
            <div className="bg-white rounded-lg border border-gray-200 shadow-sm overflow-hidden">
                {Array.from({ length: rows }).map((_, i) => (
                    <div key={i} className="flex items-center p-4 border-b border-gray-200 last:border-b-0 gap-4">
                        <div className="flex-1 space-y-2">
                            <div
                                className="h-4 bg-gray-200 rounded animate-pulse"
                                style={{ width: `${35 + (i * 17) % 40}%` }}
                            />
                            <div
                                className="h-3 bg-gray-100 rounded animate-pulse"
                                style={{ width: `${20 + (i * 11) % 25}%` }}
                            />
                        </div>
                        <div className="h-8 w-8 bg-gray-200 rounded-lg animate-pulse flex-shrink-0" />
                    </div>
                ))}
            </div>
        </div>
    );
}

export default LoadingPage
