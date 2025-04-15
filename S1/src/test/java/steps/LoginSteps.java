package steps;

import io.cucumber.java.en.*;
import org.openqa.selenium.WebDriver;
import org.openqa.selenium.chrome.ChromeDriver;
import io.github.bonigarcia.wdm.WebDriverManager;
import pages.LoginPage;
import org.testng.Assert;

public class LoginSteps {
    WebDriver driver;
    LoginPage loginPage;

    @Given("I open the browser and navigate to login page")
    public void openBrowserAndNavigate() {
       //WebDriverManager.chromedriver().setup();
        driver = new ChromeDriver();
        driver.manage().window().maximize();
        driver.get("https://demowebshop.tricentis.com/login");
        loginPage = new LoginPage(driver);
    }

    @When("I enter username {string} and password {string}")
    public void enterCredentials(String email, String password) {
        loginPage.enterEmail(email);
        loginPage.enterPassword(password);
        loginPage.clickLogin();
    }

    @Then("I should see the login status {string}")
    public void verifyLoginStatus(String expectedMessage) {
        String actual = loginPage.getLoginStatusText();
        Assert.assertTrue(actual.contains(expectedMessage));
        driver.quit();
        System.out.println("done");
    }
}
