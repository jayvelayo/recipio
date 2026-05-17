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

                    <div className="relative my-6">
                        <div className="absolute inset-0 flex items-center">
                            <div className="w-full border-t border-gray-300" />
                        </div>
                        <div className="relative flex justify-center text-sm">
                            <span className="px-2 bg-white text-gray-500">or</span>
                        </div>
                    </div>

                    <a
                        href="/auth/google"
                        className="w-full flex items-center justify-center gap-3 px-4 py-2 border border-gray-300 rounded-lg text-sm font-medium text-gray-700 hover:bg-gray-50 transition"
                    >
                        <svg viewBox="0 0 24 24" className="w-5 h-5" xmlns="http://www.w3.org/2000/svg">
                            <path d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z" fill="#4285F4"/>
                            <path d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z" fill="#34A853"/>
                            <path d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l3.66-2.84z" fill="#FBBC05"/>
                            <path d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z" fill="#EA4335"/>
                        </svg>
                        Continue with Google
                    </a>

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
