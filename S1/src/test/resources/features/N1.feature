Feature: Login to Demo Web Shop
@smoke
  Scenario Outline: Validating login with different credentials
    Given I open the browser and navigate to login page
    When I enter username "<email>" and password "<password>"
    Then I should see the login status "<expectedMessage>"

    Examples:
      | email| password   | expectedMessage |
      | nik@abc.com | nikhil1010 | nik@abc.com |
      | nik@abc.com | wrongpass  | nik@abc.com |
