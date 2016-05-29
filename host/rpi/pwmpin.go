// PWM support on the RPI.

package rpi

import (
	"fmt"
	"os"
	"strconv"

	"github.com/zlowred/embd"
	"github.com/zlowred/embd/util"
)

const (
	// PWMDefaultPolarity represents the default polarity (Positve or 1) for pwm.
	PWMDefaultPolarity = embd.Positive

	PWMDefaultDuty = 500000000

	PWMDefaultPeriod = 1000000000

	PWMMinPulseWidth = 1090
	PWMMaxPulseWidth = 1000000000
)

type pwmPin struct {
	n string

	// Either 0 or 1, depending on which pin was requested.
	// NOTE: I don't know how to tell which pin is valid, it must be set in /boot/config.txt, but I
	//       don't know how to get at that information without parsing that file.
	pwmChannel int

	drv embd.GPIODriver

	duty     int
	period   int
	polarity embd.Polarity

	enablef   *os.File
	dutyf     *os.File
	periodf   *os.File
	polarityf *os.File

	initialized bool
}

func newPWMPin(pd *embd.PinDesc, drv embd.GPIODriver) embd.PWMPin {
	pwmChannel := -1
	switch pd.DigitalLogical {
	case 12:
		pwmChannel = 0
	case 13:
		pwmChannel = 1
	default:
		panic(fmt.Sprintf("pin %d does not support pwm", pd.DigitalLogical))
	}
	return &pwmPin{n: pd.ID, drv: drv, pwmChannel: pwmChannel}
}

func (p *pwmPin) N() string {
	return p.n
}

func (p *pwmPin) id() string {
	return "rpi_pwm_" + p.n
}

func (p *pwmPin) init() error {
	if p.initialized {
		return nil
	}

	// Verify that we can read all of the relevant files.
	base := fmt.Sprintf("/sys/class/pwm/pwmchip0/pwm%d/", p.pwmChannel)

	type pairing struct {
		name string
		file **os.File
	}
	pairings := []pairing{
		{"enable", &p.enablef},
		{"duty_cycle", &p.dutyf},
		{"period", &p.periodf},
		{"polarity", &p.polarityf},
	}
	for _, pair := range pairings {
		f, err := os.OpenFile(base+pair.name, os.O_WRONLY, os.ModeExclusive)
		if err != nil {
			return fmt.Errorf("unable to open necessary pwm files: %v", err)
		}
		*pair.file = f
		defer func() {
			if !p.initialized {
				f.Close()
			}
		}()
	}

	if err := p.reset(true); err != nil {
		return err
	}
	p.initialized = true
	return nil
}

func (p *pwmPin) SetPeriod(ns int) error {
	ns /= 10
	if err := p.init(); err != nil {
		return err
	}

	if ns < PWMMinPulseWidth || ns > PWMMaxPulseWidth {
		return fmt.Errorf("embd: pwm period for %v is out of bounds, must be in [%d, %d]", p.n, PWMMinPulseWidth, PWMMaxPulseWidth)
	}

	// TODO: This might need to be scaled.
	_, err := p.periodf.WriteString(strconv.Itoa(ns))
	if err != nil {
		return err
	}

	p.period = ns

	return nil
}

func (p *pwmPin) SetDuty(ns int) error {
	ns /= 10
	if err := p.init(); err != nil {
		return err
	}

	if ns > p.period {
		return fmt.Errorf("embd: pwm duty %v for pin %v is greater than the period, %d", ns, p.n, p.period)
	}
	if ns < 0 {
		return fmt.Errorf("embd: pwm duty %v for pin %v is out of bounds (must be positive)", ns, p.n)
	}

	// TODO: This might need to be scaled
	_, err := p.dutyf.WriteString(strconv.Itoa(ns))
	if err != nil {
		return err
	}

	return nil
}

func (p *pwmPin) SetMicroseconds(us int) error {
	return fmt.Errorf("not supported (yet.  too lazy)")
	// if err := p.init(); err != nil {
	// 	return err
	// }

	// if p.period != 20000000 {
	// 	glog.Warningf("embd: pwm pin %v has freq %v hz. recommended 50 hz for servo mode", 1000000000/p.period)
	// }
	// duty := us * 1000 // in nanoseconds
	// if duty > p.period {
	// 	return fmt.Errorf("embd: calculated pwm duty %vns for pin %v (servo mode) is greater than the period %vns", duty, p.n, p.period)
	// }
	// return p.SetDuty(duty)
}

func (p *pwmPin) SetAnalog(value byte) error {
	duty := util.Map(int64(value), 0, 255, 0, int64(p.period))
	return p.SetDuty(int(duty))
}

func (p *pwmPin) SetPolarity(pol embd.Polarity) error {
	if err := p.init(); err != nil {
		return err
	}

	_, err := p.polarityf.WriteString(strconv.Itoa(int(pol)))
	if err != nil {
		return err
	}

	p.polarity = pol

	return nil
}

func (p *pwmPin) reset(enable bool) error {
	if _, err := p.polarityf.WriteString("normal\n"); err != nil {
		return err
	}
	if _, err := p.periodf.WriteString(fmt.Sprintf("%d\n", PWMDefaultPeriod/10)); err != nil {
		return err
	}
	if _, err := p.dutyf.WriteString(fmt.Sprintf("%d\n", PWMDefaultDuty/10)); err != nil {
		return err
	}
	enableStr := "0\n"
	if enable {
		enableStr = "1\n"
	}
	if _, err := p.enablef.WriteString(enableStr); err != nil {
		return err
	}
	return nil
}

func (p *pwmPin) Close() error {
	if err := p.drv.Unregister(p.n); err != nil {
		return err
	}
	if !p.initialized {
		return nil
	}
	p.initialized = false
	if err := p.reset(false); err != nil {
		return err
	}
	return nil
}
