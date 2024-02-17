import { useParams } from 'react-router-dom'
import { Profile } from '../components/Profile'
import { useUserByID } from '../hooks/useUserByID'
import { NotFound } from '../components/NotFound'
import { Navbar } from '../components/Navbar'

export function ProfilePage() {
  const { userID } = useParams()
  if (userID === undefined) {
    return <NotFound />
  }

  const { user } = useUserByID({ userID })
  return (
    <>
      <Navbar userID={user.id} />
      <Profile user={user} />
    </>
  )
}
