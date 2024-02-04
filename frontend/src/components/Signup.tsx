import { useId, useState } from 'react'
import { Female, Lock, Mail, Male, User } from './Icons'
import { Link, useNavigate } from 'react-router-dom'
import './Signup.css'

export function Signup() {
  const navigate = useNavigate()
  const [error, setError] = useState('')
  const inputUsernameID = useId()
  const inputEmailID = useId()
  const inputPasswordID = useId()
  const inputMaleID = useId()
  const inputFemaleID = useId()

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
      setError('Invalid credentials'); return
    }

    navigate('/login')
  }

  const handleSubmit = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()

    const form = e.target as HTMLFormElement
    const formData = new FormData(form)

    const username = formData.get('username') as string
    const email = formData.get('email') as string
    const password = formData.get('password') as string
    const gender = formData.get('gender') as string

    signup(username, email, password, gender)
  }

  return (
    <div className='signupContainer'>
      <div className='signup'>
        <h1 className='signup__title'>Sign Up</h1>
        <span className='signup__error'>{error}</span>
        <form className='signup__form' onSubmit={handleSubmit}>
          <div className='signup__inputField'>
            <label className='signup__label' htmlFor={inputUsernameID}>
              <User />
            </label>
            <input className='signup__input' id={inputUsernameID} name='username' type="text" placeholder='Username' />
          </div>
          <div className='signup__inputField'>
            <label className='signup__label' htmlFor={inputEmailID}>
              <Mail />
            </label>
            <input className='signup__input' id={inputEmailID} name='email' type="text" placeholder='Email' />
          </div>
          <div className='signup__inputField'>
            <label className='signup__label' htmlFor={inputPasswordID}>
              <Lock />
            </label>
            <input className='signup__input' id={inputPasswordID} name='password' type="text" placeholder='Password' />
          </div>
          <div className='signup__genderInputField'>
            <div>
              <input className='signup__inputRadio' type="radio" id={inputMaleID} name='gender' value="male" />
              <label className='signup__labelRadio' htmlFor={inputMaleID}>Male <Male /></label>
            </div>
            <div>
              <input className='signup__inputRadio' type="radio" id={inputFemaleID} name='gender' value="female" />
              <label className='signup__labelRadio' htmlFor={inputFemaleID}>Female <Female /></label>
            </div>
          </div>
          <button className='signup__submit'>Sign up</button>
        </form>
        <div className='signup__alreayHaveAnAccount'>
          <p>Do you already have an account?</p>
          <Link className='notForm' to="/login"> Log-in</Link>
        </div>
      </div>
    </div>
  )
}
