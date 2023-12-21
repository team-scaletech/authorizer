import React, { Fragment, useEffect, useState } from 'react';

import {
	AuthorizerTOTPScanner,
	AuthorizerVerifyOtp,
} from '@authorizerdev/authorizer-react';

// Define interfaces for User
interface User {
	email: string;
}

// Define interfaces for AuthResponse
interface AuthResponse {
	message: string;
	should_show_totp_screen?: boolean;
	user: User;
	authenticator_scanner_image?: string;
	authenticator_secret?: string;
	authenticator_recovery_codes?: string[];
}

// Initial data structure with default values
const initTotpData: AuthResponse = {
	message: '',
	should_show_totp_screen: false,
	user: {
		email: '',
	},
	authenticator_scanner_image: '',
	authenticator_secret: '',
};

export default function VerifyOtp() {
	console.log(`INNN`);
	
	// State variables
	const [data, setData] = useState<AuthResponse>({ ...initTotpData });
	const [loading, setLoading] = useState(false);

	// Fetch data from the server using GraphQL
	const fetchData = () => {
		setLoading(true);
		fetch('/graphql', {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json',
			},
			credentials: 'include',
			body: JSON.stringify({
				query: `{
					totp_info {
						message
						should_show_totp_screen
						authenticator_recovery_codes
						authenticator_secret
						authenticator_scanner_image
						user {
							id
							email
							email_verified
							signup_methods
							given_name
							family_name
							middle_name
							nickname
							preferred_username
							gender
							birthdate
							phone_number
							phone_number_verified
							picture
							roles
							created_at
							updated_at
							revoked_timestamp
							is_multi_factor_auth_enabled
							app_data
						}
					}
				}`,
			}),
		})
			.then(async (res) => {
				setLoading(false);
				const { data } = await res.json();
				setData(data.totp_info);
			})
			.catch((err) => {
				console.log(err);
				setLoading(false);
			});
	};

	useEffect(() => fetchData(), []);

	// If loading, display a loading message
	if (loading) {
		return <div>Loading...</div>;
	}

	// Destructure properties from the data object
	const {
		should_show_totp_screen,
		user: { email },
		authenticator_scanner_image,
		authenticator_secret,
		authenticator_recovery_codes,
	} = data;

	// Conditional rendering based on data
	if (
		authenticator_scanner_image !== '' &&
		authenticator_secret !== '' &&
		(authenticator_recovery_codes ?? []).length > 0 &&
		email != ''
	) {
		// Render TOTP scanner if conditions are met
		return (
			<Fragment>
				<h1 style={{ textAlign: 'center' }}>Verify Otp</h1>
				<br />
				<AuthorizerTOTPScanner
					email={email}
					authenticator_scanner_image={authenticator_scanner_image || ''}
					authenticator_secret={authenticator_secret || ''}
					authenticator_recovery_codes={authenticator_recovery_codes || []}
				></AuthorizerTOTPScanner>
			</Fragment>
		);
	} else {
		// Render VerifyOtp component if conditions are not met
		return (
			<Fragment>
				<h1 style={{ textAlign: 'center' }}>Verify Otp</h1>
				<br />
				<AuthorizerVerifyOtp
					email={email}
					is_totp={should_show_totp_screen}
				></AuthorizerVerifyOtp>
			</Fragment>
		);
	}
}
