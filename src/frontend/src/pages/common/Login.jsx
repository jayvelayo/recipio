import { useState } from 'react';
import { useAuth } from '/src/pages/common/AuthContext';

function Login() {
    const { login } = useAuth();
    const [formData, setFormData] = useState({
        username: '',
        password: ''
    })
    const handleFormChange = (e) => {
        const {name, value} = e.target
        setFormData({...formData, [name]: value})
    }
    const handleSubmit = (e) => {
        e.preventDefault()
        login(formData)
    }
    return (
        <>
        <h1>Log in to continue:</h1>
        <form className="ui form">
            <div class="field">
                <label>Username</label>
                <input name="username" value={formData.username} type="text" placeholder="Username" onChange={handleFormChange} required/>
            </div>
            <div class="field">
                <label>Password</label>
                <input name="password" value={formData.password} type="password" onChange={handleFormChange} required/>
            </div>
            <button className="ui primary button" type="submit" onClick={handleSubmit}>Log in</button>
        </form>
        </>
    )
}

export default Login