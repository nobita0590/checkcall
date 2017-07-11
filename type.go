package main


type (
	CallResponse struct {
		Status          string      `json:"status"`
		Message         string      `json:"message"`
		UpdateStatus    int         `json:"update_status"`
		Customer        Customer    `json:"customer"`
		Job             Job         `json:"job"`
	}
	Customer struct {
		Customer_id     string              `json:"customer_id"`
		Customer_name   string              `json:"customer_name"`
		Mobile_number   string              `json:"mobile_number"`
		Address         string              `json:"address"`
		Use_drug_start_date     string      `json:"use_drug_start_date"`
		Disease_status          string      `json:"disease_status"`
		Disease_history            string   `json:"disease_history"`
	}
	Job struct {
		Job_id          string                  `json:"job_id"`
		Job_name        string                  `json:"job_name"`
		Job_content     string                 `json:"job_content"`
		Dtatus_id       string                 `json:"status_id"`
		Next_reminder_date  string          `json:"next_reminder_date"`
		Next_reminder_content_id string     `json:"next_reminder_content_id"`
		Telephonist_id  string               `json:"telephonist_id"`
		Customer_number string              `json:"customer_number"`
		Call_type       string                    `json:"call_type"`
		Start_time      string                   `json:"start_time"`
		End_time        string                     `json:"end_time"`
		Manager_id      string                   `json:"manager_id"`
		Customer_id     string                  `json:"customer_id"`
		Job_result_id   string                `json:"job_result_id"`
	}
)
