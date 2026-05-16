
import { FiLoader } from 'react-icons/fi';

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

export default LoadingPage