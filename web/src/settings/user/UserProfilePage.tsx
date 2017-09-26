// import * as React from 'react'
// import { Link } from 'react-router-dom'
// import { Subscription } from 'rxjs/Subscription'
// import { currentUser } from '../../auth'
// import { PageTitle } from '../../components/PageTitle'
// import { TeamsTable } from '../team/TeamsTable'
// import { UserAvatar } from './UserAvatar'

// interface Props { }
// interface State {
//     user: GQL.IUser | null
//     error?: Error
// }

// /**
//  * A landing page for the user to sign in or register, if not authed
//  */
// export class UserProfilePage extends React.Component<Props, State> {
//     private subscriptions = new Subscription()

//     constructor() {
//         super()
//         this.state = {
//             user: null
//         }
//     }

//     public componentDidMount(): void {
//         this.subscriptions.add(currentUser.subscribe(
//             user => this.setState({ user }),
//             error => this.setState({ error })
//         ))
//     }

//     public componentWillUnmount(): void {
//         this.subscriptions.unsubscribe()
//     }

//     public render(): JSX.Element | null {
//         return (
//             <div className='user-profile-page'>
//                 <div className='ui-section'>
//                     <PageTitle title='profile' />
//                     <h1>Your Sourcegraph profile</h1>
//                     <div className='user-profile-page__split-row'>
//                         <div className='user-profile-page__avatar-column'>
//                             <UserAvatar size={128} onClick={() => undefined} />
//                         </div>
//                         <form className='settings-page__form'>
//                             <input readOnly type='text' className='ui-text-box'
//                                 value={this.state.user && this.state.user.email || ''} placeholder='Email' />
//                             {/* TODO(felix): make this form editable
//                             <p>
//                                 <input type='submit' className='settings-btn btn-primary btn btn-primary--right' value='Save' />
//                             </p> */}
//                         </form>
//                     </div>
//                 </div>
//                 <div className='ui-section'>
//                     <h1>Your teams</h1>
//                     <TeamsTable teams={this.state.user && this.state.user.orgs || []} />
//                     <p>
//                         <Link to='/settings/teams/new' >
//                             <input type='button' className='btn btn-primary' value='Create a new team' />
//                         </Link>
//                     </p>
//                 </div>
//             </div>
//         )
//     }
// }
