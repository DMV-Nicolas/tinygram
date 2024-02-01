import { useState, useEffect } from 'react'
import { type PostType, type UserType, type UserResponse, type LikesCountResponse } from '../types'

const DEFAULT_USER: UserType = {
  avatar: '',
  createdAt: '',
  description: '',
  email: '',
  fullName: '',
  gender: '',
  id: '',
  username: ''
}

export const usePost = ({ post }: { post: PostType }) => {
  const [user, setUser] = useState(DEFAULT_USER)
  const [likes, setLikes] = useState(0)

  useEffect(() => {
    const fetchGetUser = async () => {
      const res = await fetch(`http://localhost:5000/v1/users/${post.userID}`)
      const data: UserResponse = await res.json()
      const user: UserType = {
        id: data.id,
        username: data.username,
        fullName: data.full_name,
        email: data.email,
        avatar: data.avatar,
        description: data.description,
        gender: data.gender,
        createdAt: data.created_at
      }
      setUser(user)
    }

    fetchGetUser()
  }, [post.userID])

  useEffect(() => {
    const fetchCountLikes = async () => {
      const res = await fetch(`http://localhost:5000/v1/likes/${post.id}/count`)
      const data: LikesCountResponse = await res.json()
      setLikes(data)
    }

    fetchCountLikes()
  }, [post.id])

  return { user, likes }
}
