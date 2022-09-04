import { AuthProvider } from '../../app/auth'
import { GoogleSignIn } from './GoogleSignIn'

export const GOOGLE_API_KEY = 'AIzaSyAY47K3C3yXPRphpXbQ-vDqbu0QcjZBTNQ'
export const GOOGLE_OAUTH_CLIENT_ID = '755114359016-cr3f5lqq7d5os81br47kag73gvc4k3ud.apps.googleusercontent.com'

export const googleAuthProvider: AuthProvider = {
    name: 'Google',
    signInComponent: GoogleSignIn,
}
