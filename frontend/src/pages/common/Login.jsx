import { useState } from 'react';
import { useAuth } from '/src/pages/common/AuthContext';
import { loginAPI, getUserInfo } from '/src/pages/common/auth_apis';
import { FiMail, FiLock } from 'react-icons/fi';

function Login({ onShowRegister }) {
    const { login } = useAuth();
    const [formData, setFormData] = useState({ email: '', password: '' });
    const [error, setError] = useState('');
    const [isLoading, setIsLoading] = useState(false);

    const handleFormChange = (e) => {
        const { name, value } = e.target;
        setFormData({ ...formData, [name]: value });
    };

    const handleSubmit = async (e) => {
        e.preventDefault();
        setError('');
        setIsLoading(true);
        try {
            const { token, email } = await loginAPI(formData.email, formData.password);
            const { name } = await getUserInfo(token);
            login(token, email, name);
        } catch (err) {
            setError(err.message || 'Sign in failed');
        } finally {
            setIsLoading(false);
        }
    };

    return (
        <div className="min-h-screen bg-gradient-to-br from-indigo-600 to-indigo-800 flex items-center justify-center p-4">
            <div className="w-full max-w-md">
                <div className="bg-white rounded-lg shadow-xl p-8">
                    <div className="text-center mb-8">
                        <h1 className="text-4xl font-bold text-indigo-600 mb-2">Recipio</h1>
                        <p className="text-gray-600">Sign in to your account</p>
                    </div>

                    {error && (
                        <div className="mb-4 px-4 py-3 bg-red-50 border border-red-200 text-red-700 text-sm rounded-lg">
                            {error}
                        </div>
                    )}

                    <form onSubmit={handleSubmit} className="space-y-6">
                        <div>
                            <label htmlFor="email" className="block text-sm font-medium text-gray-900 mb-2">
                                Email
                            </label>
                            <div className="relative">
                                <FiMail className="absolute left-3 top-3 text-gray-400" size={20} />
                                <input
                                    id="email"
                                    name="email"
                                    value={formData.email}
                                    type="email"
                                    placeholder="Enter your email"
                                    onChange={handleFormChange}
                                    required
                                    className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition"
                                />
                            </div>
                        </div>

                        <div>
                            <label htmlFor="password" className="block text-sm font-medium text-gray-900 mb-2">
                                Password
                            </label>
                            <div className="relative">
                                <FiLock className="absolute left-3 top-3 text-gray-400" size={20} />
                                <input
                                    id="password"
                                    name="password"
                                    value={formData.password}
                                    type="password"
                                    placeholder="Enter your password"
                                    onChange={handleFormChange}
                                    required
                                    className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition"
                                />
                            </div>
                        </div>

                        <button
                            type="submit"
                            disabled={isLoading}
                            className="w-full bg-indigo-600 text-white font-medium py-2 rounded-lg hover:bg-indigo-700 disabled:bg-indigo-400 disabled:cursor-not-allowed transition duration-200"
                        >
                            {isLoading ? 'Signing in...' : 'Sign in'}
                        </button>
                    </form>

                    <p className="text-center text-sm text-gray-600 mt-6">
                        Don't have an account?{' '}
                        <button
                            onClick={onShowRegister}
                            className="text-indigo-600 font-medium hover:underline"
                        >
                            Create one
                        </button>
                    </p>
                </div>
            </div>
        </div>
    );
}

export default Login;
