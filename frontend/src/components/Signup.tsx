import { useId, useState } from 'react'
import { Lock, Mail, User } from './Icons'
import { Link } from 'react-router-dom'
import './Signup.css'

export function Signup() {
  const [error, setError] = useState('')
  const inputUsernameID = useId()
  const inputEmailID = useId()
  const inputPasswordID = useId()

  const signup = async (username: string, email: string, password: string, gender: string) => {
    const res = await fetch('http://localhost:5000/v1/users', {
      method: 'POST',
      body: JSON.stringify({
        username,
        email,
        password,
        gender,
        full_name: username,
        avatar: 'https://cdn-icons-png.flaticon.com/512/1068/1068549.png'
      }),
      headers: {
        'Content-Type': 'application/json'
      }
    })

    if (!res.ok) {
      setError('Invalid credentials')
      return
    }

    const data = await res.json()
    console.log(data)
  }

  const handleSubmit = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()

    const form = e.target as HTMLFormElement
    const formData = new FormData(form)

    const username = formData.get('username') as string
    const email = formData.get('email') as string
    const password = formData.get('password') as string

    signup(username, email, password, 'male')
  }

  return (
    <div className='container'>
      <div className='signup'>
        <h1 className='title'>Sign Up</h1>
        <span style={{ color: 'red' }}>{error}</span>
        <form className='form' onSubmit={handleSubmit}>
          <div className='inputField'>
            <label htmlFor={inputUsernameID}>
              <User />
            </label>
            <input id={inputUsernameID} name='username' type="text" placeholder='Username' />
          </div>
          <div className='inputField'>
            <label htmlFor={inputEmailID}>
              <Mail />
            </label>
            <input id={inputEmailID} name='email' type="text" placeholder='Email' />
          </div>
          <div className='inputField'>
            <label htmlFor={inputPasswordID}>
              <Lock />
            </label>
            <input id={inputPasswordID} name='password' type="text" placeholder='Password' />
          </div>
          <button className='submit'>Sign up</button>
        </form>
        <div className='notForm'>
          <p>Do you already have an account?</p>
          <Link className='notForm' to="/login"> Log-in</Link>
        </div>
      </div>
    </div>
  )
}
